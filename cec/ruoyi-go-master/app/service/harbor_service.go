package service

import (
	"backend/app/model"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gorm.io/gorm"
)

type HarborService struct{}

type HarborConfig struct {
	Host       string // 用于HTTP API: "http://39.96.159.232:80"
	DockerHost string // 用于Docker操作: "39.96.159.232"
	Username   string
	Password   string
}

func NewHarborConfig() *HarborConfig {
	return &HarborConfig{
		Host:       "http://39.96.159.232:80",
		DockerHost: "39.96.159.232:80",
		Username:   "admin",
		Password:   "Cecharbor147",
	}
}

// 新增：创建租户专属的 Harbor 配置
func NewTenantHarborConfig(tenantName string) *HarborConfig {
	return &HarborConfig{
		Host:       "http://39.96.159.232:80",
		DockerHost: "39.96.159.232:80",
		Username:   tenantName,
		Password:   tenantName + "Harbor2024",
	}
}

// 创建Harbor用户 - 使用用户名加固定后缀作为密码
func (s *HarborService) CreateHarborUser(username string) error {
	config := NewHarborConfig()
	url := fmt.Sprintf("%s/api/v2.0/users", config.Host)

	// 生成符合要求的密码：用户名 + Harbor2024
	password := username + "Harbor2024"

	payload := strings.NewReader(fmt.Sprintf(`{
        "username": "%s",
        "password": "%s",
        "realname": "%s",
        "email": "%s@tenant.local"
    }`, username, password, username, username))

	req, _ := http.NewRequest("POST", url, payload)
	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("创建Harbor用户失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("创建Harbor用户失败: %s", string(body))
	}

	return nil
}

// 创建项目 - 使用租户配置
func (s *HarborService) CreateProject(username string) error {
	config := NewTenantHarborConfig(username)
	url := fmt.Sprintf("%s/api/v2.0/projects", config.Host)
	payload := strings.NewReader(fmt.Sprintf(`{
        "project_name": "%s",
        "metadata": {
            "public": "false"
        }
    }`, username))

	req, _ := http.NewRequest("POST", url, payload)
	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("创建Harbor项目失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("创建Harbor项目失败: %s", string(body))
	}

	return nil
}

// 获取所有基础镜像
func (s *HarborService) ListBaseImages() ([]map[string]interface{}, error) {
	config := NewHarborConfig()
	url := fmt.Sprintf("%s/api/v2.0/projects/base/repositories", config.Host)

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(config.Username, config.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取基础镜像列表失败: %v", err)
	}
	defer resp.Body.Close()

	var repositories []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("解析基础镜像列表失败: %v", err)
	}

	// 获取每个仓库的标签信息
	for i, repo := range repositories {
		repoName, ok := repo["name"].(string)
		if !ok {
			continue
		}

		// 获取该仓库的 artifacts（包含标签信息）
		artifactsURL := fmt.Sprintf("%s/api/v2.0/projects/base/repositories/%s/artifacts",
			config.Host,
			strings.TrimPrefix(repoName, "base/"))

		req, _ := http.NewRequest("GET", artifactsURL, nil)
		req.SetBasicAuth(config.Username, config.Password)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var artifacts []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
			continue
		}

		// 收集所有标签
		var tagNames []string
		for _, artifact := range artifacts {
			if tags, ok := artifact["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagInfo, ok := tag.(map[string]interface{}); ok {
						if tagName, ok := tagInfo["name"].(string); ok {
							tagNames = append(tagNames, tagName)
						}
					}
				}
			}
		}

		// 将标签列表添加到仓库信息中
		repositories[i]["tags"] = tagNames
	}

	return repositories, nil
}

// // 获取所有基础镜像
// func (s *HarborService) ListBaseImages() ([]map[string]interface{}, error) {
// 	config := NewHarborConfig()
// 	url := fmt.Sprintf("%s/api/v2.0/projects/base/repositories", config.Host)

// 	req, _ := http.NewRequest("GET", url, nil)
// 	req.SetBasicAuth(config.Username, config.Password)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("获取基础镜像列表失败: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	var images []map[string]interface{}
// 	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
// 		return nil, fmt.Errorf("解析基础镜像列表失败: %v", err)
// 	}

// 	return images, nil
// }

func (s *HarborService) PushImage(tenantName, imageName, tag string) error {
	config := NewTenantHarborConfig(tenantName)
	fmt.Printf("[PushImage] DockerHost: %s, tenantName: %s, imageName: %s, tag: %s\n",
		config.DockerHost, tenantName, imageName, tag)
	// 构建推送地址
	pushUrl := fmt.Sprintf("%s/%s/%s:%s", config.DockerHost, tenantName, imageName, tag)
	fmt.Printf("[PushImage] 构建的推送地址: %s\n", pushUrl)
	sourceImage := fmt.Sprintf("%s:%s", imageName, tag)
	// 设置环境变量
	os.Setenv("DOCKER_DEBUG", "1")
	// 创建Docker客户端，添加调试选项
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.41"), // 指定 API 版本
	)

	if err != nil {
		return fmt.Errorf("创建Docker客户端失败: %v", err)
	}
	defer cli.Close()
	// 打印调试信息
	fmt.Printf("Docker 客户端配置:\n")
	fmt.Printf("推送地址: %s\n", pushUrl)
	fmt.Printf("源镜像: %s\n", sourceImage)
	fmt.Printf("认证信息: 用户名=%s, 服务器=%s\n", config.Username, config.DockerHost)
	// 标记镜像
	err = cli.ImageTag(context.Background(), sourceImage, pushUrl)
	if err != nil {
		return fmt.Errorf("标记镜像失败: %v", err)
	}
	// 打印认证信息
	fmt.Printf("认证配置信息:\n用户名: %s\n服务器地址: %s\n", config.Username, config.DockerHost)
	// 设置认证信息
	authConfig := types.AuthConfig{
		Username:      config.Username,
		Password:      config.Password,
		ServerAddress: config.DockerHost,
	}
	// 先转为 JSON
	jsonData, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("编码认证信息失败: %v", err)
	}
	// 再进行 base64 编码
	encodedAuth := base64.URLEncoding.EncodeToString(jsonData)
	fmt.Printf("编码后的认证信息: %s\n", encodedAuth)

	// 推送镜像
	ctx := context.Background()
	opts := types.ImagePushOptions{
		RegistryAuth: string(encodedAuth),
	}
	pushResp, err := cli.ImagePush(ctx, pushUrl, opts)
	if err != nil {
		return fmt.Errorf("推送镜像失败: %v", err)
	}
	defer pushResp.Close()

	// 读取并显示推送进度
	d := json.NewDecoder(pushResp)
	var status map[string]interface{}
	for {
		if err := d.Decode(&status); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取推送进度失败: %v", err)
		}
		fmt.Printf("推送状态: %v\n", status)
	}

	return nil
}

// 拉取镜像
func (s *HarborService) PullImage(repository, tag string) error {
	fmt.Println("ddd")
	config := NewHarborConfig()
	// 构建拉取地址
	pullUrl := fmt.Sprintf("%s/%s:%s", config.DockerHost, repository, tag)
	fmt.Println("ccc")
	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("创建Docker客户端失败: %v", err)
	}
	defer cli.Close()

	// 设置认证信息
	authConfig := types.AuthConfig{
		Username:      config.Username,
		Password:      config.Password,
		ServerAddress: config.Host,
	}
	// 先转为 JSON
	jsonData, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("编码认证信息失败: %v", err)
	}
	// 再进行 base64 编码
	encodedAuth := base64.URLEncoding.EncodeToString(jsonData)
	// 拉取镜像
	ctx := context.Background()
	opts := types.ImagePullOptions{
		RegistryAuth: string(encodedAuth),
	}
	fmt.Println("bbb")
	pullResp, err := cli.ImagePull(ctx, pullUrl, opts)
	fmt.Println("ppp")
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %v", err)
	}
	defer pullResp.Close()
	// 读取并显示拉取进度
	d := json.NewDecoder(pullResp)
	var status map[string]interface{}
	for {
		if err := d.Decode(&status); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取拉取进度失败: %v", err)
		}
		// 可以打印状态信息
		fmt.Printf("拉取状态: %v\n", status)
	}
	return nil
}

// PullImageWithAuth 使用指定账号拉取镜像，repository是"repository": "harbor123/alpine",  // 私有仓库路径，以用户名开头
func (s *HarborService) PullImageWithAuth(username, repository, tag string) error {
	config := NewTenantHarborConfig(username)
	// 构建拉取地址
	pullUrl := fmt.Sprintf("%s/%s:%s", config.DockerHost, repository, tag)

	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("创建Docker客户端失败: %v", err)
	}
	defer cli.Close()

	// 设置认证信息
	authConfig := types.AuthConfig{
		Username:      config.Username,
		Password:      config.Password,
		ServerAddress: config.Host,
	}
	// 先转为 JSON
	jsonData, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("编码认证信息失败: %v", err)
	}
	// 再进行 base64 编码
	encodedAuth := base64.URLEncoding.EncodeToString(jsonData)

	// 拉取镜像
	ctx := context.Background()
	opts := types.ImagePullOptions{
		RegistryAuth: string(encodedAuth),
	}

	pullResp, err := cli.ImagePull(ctx, pullUrl, opts)
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %v", err)
	}
	defer pullResp.Close()

	// 读取并显示拉取进度
	d := json.NewDecoder(pullResp)
	var status map[string]interface{}
	for {
		if err := d.Decode(&status); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取拉取进度失败: %v", err)
		}
		// 可以打印状态信息
		fmt.Printf("拉取状态: %v\n", status)
	}
	return nil
}

// ... existing code ...

// 异步推送镜像
func (s *HarborService) AsyncPushImage(db *gorm.DB, task *model.ImageTask) error {
	go func() {
		// 开启事务
		tx := db.Begin()
		if tx.Error != nil {
			fmt.Printf("开启事务失败: %v\n", tx.Error)
			return
		}

		// 更新任务状态为运行中
		task.Status = model.TaskStatusRunning
		if err := tx.Save(task).Error; err != nil {
			tx.Rollback()
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("更新任务状态失败: %v", err)
			db.Save(task)
			return
		}

		err := s.PushImage(task.Username, task.ImageName, task.Tag)
		if err != nil {
			fmt.Printf("[AsyncPushImage] 推送镜像失败: %v\n", err)
			tx.Rollback()
			// 使用新的数据库连接保存失败状态
			if err := db.Model(task).Updates(map[string]interface{}{
				"status":  model.TaskStatusFailed,
				"message": err.Error(),
			}).Error; err != nil {
				fmt.Printf("保存失败状态失败: %v\n", err)
			}
		} else {
			task.Status = model.TaskStatusComplete
			task.Message = "推送成功"
		}

		if err := tx.Save(task).Error; err != nil {
			tx.Rollback()
			fmt.Printf("保存任务状态失败: %v\n", err)
			return
		}

		// 提交事务
		tx.Commit()
	}()
	return nil
}

// 异步拉取镜像
func (s *HarborService) AsyncPullImage(db *gorm.DB, task *model.ImageTask) error {
	go func() {
		// 开启事务
		tx := db.Begin()
		if tx.Error != nil {
			fmt.Printf("开启事务失败: %v\n", tx.Error)
			return
		}

		// 更新任务状态为运行中
		task.Status = model.TaskStatusRunning
		if err := tx.Save(task).Error; err != nil {
			tx.Rollback()
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("更新任务状态失败: %v", err)
			db.Save(task)
			return
		}

		err := s.PullImageWithAuth(task.Username, task.ImageName, task.Tag)
		if err != nil {
			tx.Rollback()
			// 使用新的数据库连接保存失败状态
			if err := db.Model(task).Updates(map[string]interface{}{
				"status":  model.TaskStatusFailed,
				"message": err.Error(),
			}).Error; err != nil {
				fmt.Printf("保存失败状态失败: %v\n", err)
			}
		} else {
			task.Status = model.TaskStatusComplete
			task.Message = "拉取成功"
			fmt.Printf("拉取成功")
		}

		if err := tx.Save(task).Error; err != nil {
			tx.Rollback()
			fmt.Printf("保存任务状态失败: %v\n", err)
			return
		}

		// 提交事务
		tx.Commit()
	}()
	return nil
}

// 获取任务状态
func (s *HarborService) GetTaskStatus(db *gorm.DB, taskId int) (*model.ImageTask, error) {
	var task model.ImageTask
	if err := db.First(&task, taskId).Error; err != nil {
		return nil, fmt.Errorf("获取任务状态失败: %v", err)
	}
	return &task, nil
}

// 获取用户所有任务
func (s *HarborService) ListUserTasks(db *gorm.DB, tenantId int) ([]model.ImageTask, error) {
	var tasks []model.ImageTask
	if err := db.Where("tenant_id = ?", tenantId).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("获取用户任务列表失败: %v", err)
	}
	return tasks, nil
}

// LoadImage 加载镜像文件并推送到Harbor
func (s *HarborService) LoadImage(db *gorm.DB, task *model.ImageLoadTask) error {
	go func() {
		// 更新任务状态为运行中
		task.Status = model.TaskStatusRunning
		db.Save(task)

		// 创建Docker客户端
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("创建Docker客户端失败: %v", err)
			db.Save(task)
			return
		}
		defer cli.Close()

		// 打开tar文件
		file, err := os.Open(task.FilePath)
		if err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("打开镜像文件失败: %v", err)
			db.Save(task)
			return
		}
		defer file.Close()

		// 加载镜像
		loadResp, err := cli.ImageLoad(context.Background(), file, true)
		if err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("加载镜像失败: %v", err)
			db.Save(task)
			return
		}
		defer loadResp.Body.Close()

		// 读取加载结果
		content, err := ioutil.ReadAll(loadResp.Body)
		if err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("读取加载结果失败: %v", err)
			db.Save(task)
			return
		}
		// 记录加载结果到任务消息
		task.Message = fmt.Sprintf("镜像加载成功: %s", string(content))
		fmt.Printf("镜像加载成功: %s", string(content))
		// 解析加载结果获取镜像ID
		var result struct {
			Stream string `json:"stream"`
		}
		if err := json.Unmarshal(content, &result); err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("解析加载结果失败: %v", err)
			return
		}

		// 从结果中提取镜像ID
		imageID := strings.TrimPrefix(strings.TrimSpace(result.Stream), "Loaded image ID: ")
		imageID = strings.TrimSpace(imageID)
		// 获取Harbor配置
		config := NewTenantHarborConfig(task.Username)
		targetTag := fmt.Sprintf("%s/%s/%s:%s", config.DockerHost, task.Username, task.ImageName, task.Tag)
		fmt.Printf("ImageName: %s, TargetTag: %s\n", task.ImageName, targetTag)
		fmt.Printf("ImageName: %s, TargetTag: %s\n", task.ImageName, targetTag)

		// 使用镜像ID进行标记
		if err := cli.ImageTag(context.Background(), imageID, targetTag); err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("标记镜像失败: %v", err)
			return
		}

		// 推送镜像到Harbor
		authConfig := types.AuthConfig{
			Username:      config.Username,
			Password:      config.Password,
			ServerAddress: config.Host,
		}
		authBytes, _ := json.Marshal(authConfig)
		authStr := base64.URLEncoding.EncodeToString(authBytes)

		pushResp, err := cli.ImagePush(context.Background(), targetTag, types.ImagePushOptions{
			RegistryAuth: authStr,
		})
		if err != nil {
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("推送镜像失败: %v", err)
			db.Save(task)
			return
		}
		defer pushResp.Close()

		// 读取推送进度
		d := json.NewDecoder(pushResp)
		var status map[string]interface{}
		for {
			if err := d.Decode(&status); err != nil {
				if err == io.EOF {
					break
				}
				task.Status = model.TaskStatusFailed
				task.Message = fmt.Sprintf("读取推送进度失败: %v", err)
				db.Save(task)
				return
			}
			// 打印推送进度
			fmt.Printf("推送进度: %+v\n", status)
		}

		// 删除临时文件
		os.Remove(task.FilePath)

		// 更新任务状态为完成
		task.Status = model.TaskStatusComplete
		task.Message = "镜像加载并推送成功"
		db.Save(task)
	}()

	return nil
}

// DeleteProject 删除Harbor项目和用户账号
func (s *HarborService) DeleteProject(projectName string) error {
	// 使用租户的配置删除项目
	tenantConfig := NewTenantHarborConfig(projectName)
	projectURL := fmt.Sprintf("%s/api/v2.0/projects/%s", tenantConfig.Host, projectName)
	req, _ := http.NewRequest("DELETE", projectURL, nil)
	req.SetBasicAuth(tenantConfig.Username, tenantConfig.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("删除Harbor项目失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("删除Harbor项目失败: %s", string(body))
	}

	// 使用管理员配置删除用户账号
	config := NewHarborConfig()
	userURL := fmt.Sprintf("%s/api/v2.0/users/search?username=%s", config.Host, projectName)
	req, _ = http.NewRequest("GET", userURL, nil)
	req.SetBasicAuth(config.Username, config.Password)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("查找Harbor用户失败: %v", err)
	}
	defer resp.Body.Close()

	// 直接解析为用户数组
	var users []struct {
		UserID int `json:"user_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return fmt.Errorf("解析用户信息失败: %v", err)
	}

	if len(users) > 0 {
		// 删除找到的用户
		userDeleteURL := fmt.Sprintf("%s/api/v2.0/users/%d", config.Host, users[0].UserID)
		req, _ = http.NewRequest("DELETE", userDeleteURL, nil)
		req.SetBasicAuth(config.Username, config.Password)

		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("删除Harbor用户失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("删除Harbor用户失败: %s", string(body))
		}
	}

	return nil
}

// 获取所有项目及其镜像列表
func (s *HarborService) ListAllImages() ([]map[string]interface{}, error) {
	config := NewHarborConfig()

	// 1. 首先获取所有项目
	projectsURL := fmt.Sprintf("%s/api/v2.0/projects", config.Host)
	req, _ := http.NewRequest("GET", projectsURL, nil)
	req.SetBasicAuth(config.Username, config.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %v", err)
	}
	defer resp.Body.Close()

	var projects []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("解析项目列表失败: %v", err)
	}

	// 2. 对每个项目获取其镜像列表
	var result []map[string]interface{}
	for _, project := range projects {
		projectName, ok := project["name"].(string)
		if !ok {
			continue
		}

		// 获取该项目下的镜像
		repoURL := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories", config.Host, projectName)
		req, _ := http.NewRequest("GET", repoURL, nil)
		req.SetBasicAuth(config.Username, config.Password)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var repositories []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
			continue
		}

		// 3. 对每个仓库获取其标签列表
		for i, repo := range repositories {
			repoName, ok := repo["name"].(string)
			if !ok {
				continue
			}

			// 获取该仓库的 artifacts（包含标签信息）
			artifactsURL := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts",
				config.Host,
				projectName,
				strings.TrimPrefix(repoName, projectName+"/"))

			req, _ := http.NewRequest("GET", artifactsURL, nil)
			req.SetBasicAuth(config.Username, config.Password)

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			var artifacts []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
				continue
			}

			// 收集所有标签
			var tagNames []string
			for _, artifact := range artifacts {
				if tags, ok := artifact["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagInfo, ok := tag.(map[string]interface{}); ok {
							if tagName, ok := tagInfo["name"].(string); ok {
								tagNames = append(tagNames, tagName)
							}
						}
					}
				}
			}

			// 将标签列表添加到仓库信息中
			repositories[i]["tags"] = tagNames
		}

		// 将项目信息和镜像列表组合
		result = append(result, map[string]interface{}{
			"project_name": projectName,
			"images":       repositories,
		})
	}

	return result, nil
}
