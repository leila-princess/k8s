package service

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"backend/app/dto"
	"backend/app/model"
	"backend/config"
	"backend/framework/redis"

	"gorm.io/gorm"
)

type TenantService struct{}

type Claims struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	jwt.StandardClaims
}

func (s *TenantService) CreateTenant(db *gorm.DB, username, password string, quota model.K8sResourceQuota) (string, error) {
	// 检查用户名是否已存在
	var count int64
	if err := db.Model(&model.TenantUser{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return "", err
	}
	if count > 0 {
		return "", errors.New("用户名已存在")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	var token string
	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建租户用户
		user := &model.TenantUser{
			Username: username,
			Password: string(hashedPassword),
			Status:   "0",
			IsAdmin:  "0",
		}
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("创建租户用户失败：%v", err)
		}

		// 创建Harbor用户（在创建项目之前）
		harborService := &HarborService{}
		if err := harborService.CreateHarborUser(username); err != nil {
			return fmt.Errorf("创建Harbor用户失败：%v", err)
		}

		// 创建Harbor项目（使用租户自己的账号）
		if err := harborService.CreateProject(username); err != nil {
			return fmt.Errorf("创建Harbor项目失败：%v", err)
		}

		// 创建命名空间记录
		namespace := &model.K8sNamespace{
			Name:   username,
			UserId: user.Id,
			Status: "0",
		}
		if err := tx.Create(namespace).Error; err != nil {
			return fmt.Errorf("创建命名空间记录失败：%v", err)
		}

		// 创建资源配额记录
		quota.NamespaceId = namespace.Id
		if err := tx.Create(&quota).Error; err != nil {
			return fmt.Errorf("创建资源配额记录失败：%v", err)
		}

		// 创建Kubernetes命名空间
		namespaceService := &K8sNamespaceService{}
		if err := namespaceService.CreateNamespace(username); err != nil {
			return fmt.Errorf("创建Kubernetes命名空间失败：%v", err)
		}

		// 创建默认的 LimitRange
		if err := namespaceService.CreateDefaultLimitRange(username); err != nil {
			return fmt.Errorf("创建默认LimitRange失败：%v", err)
		}

		// 创建Kubernetes资源配额
		if err := namespaceService.CreateResourceQuota(username, quota); err != nil {
			return fmt.Errorf("创建Kubernetes资源配额失败：%v", err)
		}

		// 创建ServiceAccount
		rbacService := &K8sRbacService{}
		serviceAccountName := username + "-sa"
		if err := rbacService.CreateServiceAccount(username, serviceAccountName); err != nil {
			return fmt.Errorf("创建ServiceAccount失败：%v", err)
		}

		// 更新命名空间记录的ServiceAccount信息
		if err := tx.Model(namespace).Update("service_account", serviceAccountName).Error; err != nil {
			return fmt.Errorf("更新ServiceAccount信息失败：%v", err)
		}

		// 创建Role和RoleBinding
		if err := rbacService.CreateNamespaceRole(username); err != nil {
			return fmt.Errorf("创建Role失败：%v", err)
		}

		if err := rbacService.CreateRoleBinding(username, serviceAccountName); err != nil {
			return fmt.Errorf("创建RoleBinding失败：%v", err)
		}

		// 获取初始token并缓存到Redis
		token, err = rbacService.GetServiceAccountToken(username, serviceAccountName)
		if err != nil {
			return fmt.Errorf("获取ServiceAccount token失败：%v", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	return token, nil
}

// Login 用户登录
func (s *TenantService) Login(db *gorm.DB, username, password string) (string, model.TenantUser, error) {
	var user model.TenantUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", model.TenantUser{}, errors.New("用户不存在")
		}
		return "", model.TenantUser{}, err
	}

	// 检查密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", model.TenantUser{}, errors.New("密码错误")
	}

	// 检查用户状态
	if user.Status == "1" {
		return "", model.TenantUser{}, errors.New("账户已停用")
	}

	// 生成Token
	claims := Claims{
		UserId:   user.Id,
		Username: user.Username,
		IsAdmin:  user.IsAdmin == "1",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(config.TokenExpireDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", model.TenantUser{}, err
	}

	// 将token存入Redis
	key := fmt.Sprintf("token:%d", user.Id)
	if err := redis.Set(key, tokenString, config.TokenExpireDuration); err != nil {
		return "", model.TenantUser{}, err
	}

	// 清除密码字段
	user.Password = ""

	return tokenString, user, nil
}

// UpdateTenantById 通过ID更新租户信息
func (s *TenantService) UpdateTenantById(db *gorm.DB, req dto.UpdateTenantRequest) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var user model.TenantUser
		if err := tx.First(&user, req.Id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		// 更新用户信息
		updates := map[string]interface{}{}
		if req.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			updates["password"] = string(hashedPassword)
		}
		if req.Status != "" {
			updates["status"] = req.Status
			// 同步更新命名空间状态
			var namespace model.K8sNamespace
			if err := tx.Where("user_id = ?", user.Id).First(&namespace).Error; err != nil {
				return err
			}
			if err := tx.Model(&namespace).Update("status", req.Status).Error; err != nil {
				return err
			}

			// 同步更新资源配额状态
			var quota model.K8sResourceQuota
			if err := tx.Where("namespace_id = ?", namespace.Id).First(&quota).Error; err != nil {
				return err
			}
			if err := tx.Model(&quota).Update("status", req.Status).Error; err != nil {
				return err
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateTenantByUsername 通过用户名更新租户信息
func (s *TenantService) UpdateTenantByUsername(db *gorm.DB, req dto.UpdateTenantRequest) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var user model.TenantUser
		if err := tx.Where("username = ?", req.Username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		// 更新用户信息
		updates := map[string]interface{}{}
		if req.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			updates["password"] = string(hashedPassword)
		}
		if req.Status != "" {
			updates["status"] = req.Status
			// 同步更新命名空间状态
			var namespace model.K8sNamespace
			if err := tx.Where("user_id = ?", user.Id).First(&namespace).Error; err != nil {
				return err
			}
			if err := tx.Model(&namespace).Update("status", req.Status).Error; err != nil {
				return err
			}

			// 同步更新资源配额状态
			var quota model.K8sResourceQuota
			if err := tx.Where("namespace_id = ?", namespace.Id).First(&quota).Error; err != nil {
				return err
			}
			if err := tx.Model(&quota).Update("status", req.Status).Error; err != nil {
				return err
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteTenant 删除租户
func (s *TenantService) DeleteTenant(db *gorm.DB, id int) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var user model.TenantUser
		if err := tx.First(&user, id).Error; err != nil {
			return err
		}
        // 删除Harbor项目
		harborService := &HarborService{}
		if err := harborService.DeleteProject(user.Username); err != nil {
			return fmt.Errorf("删除Harbor项目失败：%v", err)
		}
		// 删除命名空间
		var namespace model.K8sNamespace
		if err := tx.Where("user_id = ?", user.Id).First(&namespace).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		} else {
			// 删除K8s中的命名空间（这将自动删除相关的资源配额）
			client, err := (&K8sConfigService{}).GetClient()
			if err != nil {
				return err
			}
			if err := client.CoreV1().Namespaces().Delete(context.TODO(), namespace.Name, metav1.DeleteOptions{}); err != nil {
				if !k8serrors.IsNotFound(err) {
					return err
				}
			}

			// 删除资源配额记录
			if err := tx.Where("namespace_id = ?", namespace.Id).Delete(&model.K8sResourceQuota{}).Error; err != nil {
				return err
			}

			// 删除命名空间记录
			if err := tx.Delete(&namespace).Error; err != nil {
				return err
			}
		}

		// 删除用户
		if err := tx.Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteTenantByUsername 通过用户名删除租户
func (s *TenantService) DeleteTenantByUsername(db *gorm.DB, username string) error {
	var user model.TenantUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	return s.DeleteTenant(db, user.Id)
}

// UpdateTenantQuotaById 通过ID更新租户资源配额
func (s *TenantService) UpdateTenantQuotaById(db *gorm.DB, id int, newQuota model.K8sResourceQuota) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 查找用户
		var user model.TenantUser
		if err := tx.First(&user, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		// 查找命名空间
		var namespace model.K8sNamespace
		if err := tx.Where("user_id = ?", user.Id).First(&namespace).Error; err != nil {
			return err
		}

		// 查找现有资源配额
		var quota model.K8sResourceQuota
		if err := tx.Where("namespace_id = ?", namespace.Id).First(&quota).Error; err != nil {
			return err
		}

		// 更新数据库中的资源配额
		newQuota.Id = quota.Id                   // 保持ID不变
		newQuota.NamespaceId = quota.NamespaceId // 保持NamespaceId不变
		newQuota.Status = quota.Status           // 保持状态不变
		if err := tx.Model(&quota).Updates(newQuota).Error; err != nil {
			return err
		}

		// 更新K8s中的资源配额
		namespaceService := &K8sNamespaceService{}
		if err := namespaceService.UpdateResourceQuota(namespace.Name, newQuota); err != nil {
			return fmt.Errorf("更新K8s资源配额失败：%v", err)
		}

		return nil
	})
}

// UpdateTenantQuotaByUsername 通过用户名更新租户资源配额
func (s *TenantService) UpdateTenantQuotaByUsername(db *gorm.DB, username string, newQuota model.K8sResourceQuota) error {
	var user model.TenantUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	return s.UpdateTenantQuotaById(db, user.Id, newQuota)
}

// UpdatePassword 更新用户密码
func (s *TenantService) UpdatePassword(db *gorm.DB, userId int, newPassword string) error {
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	if err := db.Model(&model.TenantUser{}).Where("id = ?", userId).Update("password", string(hashedPassword)).Error; err != nil {
		return err
	}

	return nil
}

// Logout 退出登录
func (s *TenantService) Logout(db *gorm.DB, userId int) error {
	// 从Redis中删除token
	key := fmt.Sprintf("token:%d", userId)
	if err := redis.Del(key); err != nil {
		return fmt.Errorf("清除token失败：%v", err)
	}
	return nil
}

// ListUsers 获取所有用户信息
func (s *TenantService) ListUsers(db *gorm.DB) ([]model.TenantUser, error) {
	var users []model.TenantUser
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	// 清除所有用户的密码字段
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}
