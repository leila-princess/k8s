package service

import (
	"backend/framework/redis"
	"context"
	"fmt"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type K8sRbacService struct{}

// 创建集群角色（用于管理员）
func (s *K8sRbacService) CreateAdminClusterRole() error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ruoyi-admin",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	_, err = client.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
	return err
}

// 创建集群角色绑定（用于管理员）
func (s *K8sRbacService) CreateAdminClusterRoleBinding(username string) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ruoyi-admin-binding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     username,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "ruoyi-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = client.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	return err
}

// 创建命名空间角色
func (s *K8sRbacService) CreateNamespaceRole(namespace string) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace + "-admin",
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	_, err = client.RbacV1().Roles(namespace).Create(context.TODO(), role, metav1.CreateOptions{})
	return err
}

// 创建角色绑定
func (s *K8sRbacService) CreateRoleBinding(namespace, serviceAccount string) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace + "-admin-binding",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     namespace + "-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = client.RbacV1().RoleBindings(namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
	return err
}

// 创建ServiceAccount
func (s *K8sRbacService) CreateServiceAccount(namespace, name string) error {
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return err
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err = client.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
	return err
}

// GetServiceAccountToken 获取ServiceAccount的token，优先从Redis获取，不存在则从K8s获取
func (s *K8sRbacService) GetServiceAccountToken(namespace, serviceAccountName string) (string, error) {
	// 构造Redis key
	redisKey := fmt.Sprintf("k8s:sa:token:%s:%s", namespace, serviceAccountName)

	// 尝试从Redis获取token
	token, err := redis.Get(redisKey)
	if err == nil && token != "" {
		return token, nil
	}

	// Redis中不存在，从K8s获取
	client, err := (&K8sConfigService{}).GetClient()
	if err != nil {
		return "", err
	}

	// 使用 TokenRequest API 获取短期token
	tokenRequest := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			//Audiences:         []string{"https://kubernetes.default.svc"}, // API服务器的默认audience
			ExpirationSeconds: pointer.Int64(3600), // token有效期1小时
		},
	}

	// 调用 TokenRequest API
	tr, err := client.CoreV1().ServiceAccounts(namespace).CreateToken(context.TODO(),
		serviceAccountName, tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create token: %v", err)
	}

	// 将token存入Redis，设置1小时过期
	if err := redis.Set(redisKey, tr.Status.Token, time.Hour); err != nil {
		// 仅记录错误，不影响返回
		fmt.Printf("failed to cache token in redis: %v\n", err)
	}

	return tr.Status.Token, nil
}

// // 获取ServiceAccount的token
// func (s *K8sRbacService) GetServiceAccountToken(namespace, serviceAccountName string) (string, error) {
// 	client, err := (&K8sConfigService{}).GetClient()
// 	if err != nil {
// 		return "", err
// 	}

// 	// 使用 TokenRequest API 获取短期token
// 	tokenRequest := &authenticationv1.TokenRequest{
// 		Spec: authenticationv1.TokenRequestSpec{
// 			Audiences:         []string{"https://kubernetes.default.svc"}, // API服务器的默认audience
// 			ExpirationSeconds: pointer.Int64(3600),                        // token有效期1小时
// 		},
// 	}

// 	// 调用 TokenRequest API
// 	tr, err := client.CoreV1().ServiceAccounts(namespace).CreateToken(context.TODO(),
// 		serviceAccountName, tokenRequest, metav1.CreateOptions{})
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create token: %v", err)
// 	}

// 	return tr.Status.Token, nil
// }
