package service

import (
	"backend/app/dto"
	"backend/app/model"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// K8sDeployment Kubernetes部署配置
type K8sDeployment struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels,omitempty"`
	} `yaml:"metadata"`
	Spec struct {
		Replicas int32 `yaml:"replicas"`
		Selector struct {
			MatchLabels map[string]string `yaml:"matchLabels"`
		} `yaml:"selector"`
		Template struct {
			Metadata struct {
				Labels map[string]string `yaml:"labels"`
			} `yaml:"metadata"`
			Spec struct {
				Containers []struct {
					Name    string              `yaml:"name"`
					Image   string              `yaml:"image"`
					Command []string            `yaml:"command,omitempty"`
					Args    []string            `yaml:"args,omitempty"`
					Ports   []dto.ContainerPort `yaml:"ports,omitempty"`
					Env     []struct {
						Name  string `yaml:"name"`
						Value string `yaml:"value"`
					} `yaml:"env,omitempty"`
					Resources struct {
						Requests struct {
							CPU    string `yaml:"cpu,omitempty"`
							Memory string `yaml:"memory,omitempty"`
						} `yaml:"requests,omitempty"`
						Limits struct {
							CPU    string `yaml:"cpu,omitempty"`
							Memory string `yaml:"memory,omitempty"`
						} `yaml:"limits,omitempty"`
					} `yaml:"resources,omitempty"`
				} `yaml:"containers"`
			} `yaml:"spec"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

func (s *ArgoCDService) getTenantId(db *gorm.DB, username string) (int, error) {
	var user model.TenantUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return 0, fmt.Errorf("获取租户信息失败: %v", err)
	}
	return user.Id, nil
}

type GitlabConfig struct {
	BaseURL string
	Token   string
}

type ArgoCDConfig struct {
	BaseURL  string
	Username string
	Password string
}

type ArgoCDService struct {
	k8sConfig    *K8sConfigService
	db           *gorm.DB
	gitlabConfig *GitlabConfig
	argoCDConfig *ArgoCDConfig
}

func NewArgoCDService(db *gorm.DB, gitlabBaseURL, gitlabToken string) *ArgoCDService {
	return &ArgoCDService{
		k8sConfig: &K8sConfigService{},
		db:        db,
		gitlabConfig: &GitlabConfig{
			BaseURL: gitlabBaseURL,
			Token:   gitlabToken,
		},
		argoCDConfig: &ArgoCDConfig{
			BaseURL:  "http://localhost:8000", // 本地开发时使用
			Username: "admin",
			Password: "Cecargocd147",
		},
	}
}

type GitlabCommitAction struct {
	Action   string `json:"action"`
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

type GitlabCommitPayload struct {
	Branch        string               `json:"branch"`
	CommitMessage string               `json:"commit_message"`
	Actions       []GitlabCommitAction `json:"actions"`
}

func extractProjectIDFromURL(repoURL string) string {
	// 移除 .git 后缀（如果存在）
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// 从 URL 中提取项目路径
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return ""
	}

	// 获取最后两部分作为项目路径（group/project）
	projectPath := strings.Join(parts[len(parts)-2:], "/")

	// URL 编码项目路径
	return url.PathEscape(projectPath)
}

// CreateArgoCDDeployment 创建ArgoCD部署
func (s *ArgoCDService) CreateArgoCDDeployment(ctx context.Context, username string, req *dto.ArgoCDDeployRequest) (*dto.ArgoCDDeployResponse, error) {
	// 首先检查接收器是否为 nil
	if s == nil {
		fmt.Println("❌ 致命错误: ArgoCDService 接收器是 nil!")
		return nil, fmt.Errorf("服务未正确初始化")
	}

	fmt.Println("✅ 服务接收器检查通过")
	// 构建标准化的YAML文件路径
	yamlPath := filepath.Join("deployments", username, req.Name+".yaml")
	// 确保使用Unix风格的路径（GitLab需要）
	yamlPath = filepath.ToSlash(yamlPath)
	req.YamlPath = yamlPath

	// 在数据库中记录部署信息
	tenantId, err := s.getTenantId(s.db, username)
	if err != nil {
		return nil, err
	}
	deployment := &model.ArgoCDDeployment{
		TenantId:    tenantId,
		Name:        req.Name,
		GitRepo:     req.GitRepo,
		GitBranch:   req.GitBranch,
		YamlPath:    yamlPath,
		Image:       req.Image,
		TriggerType: req.TriggerType,
		Schedule:    req.Schedule,
		Status:      "pending",
	}
	if req.Tag != 1 {
		// 在数据库中记录部署信息
		if err := s.db.Create(deployment).Error; err != nil {
			return nil, fmt.Errorf("保存部署信息失败: %v", err)
		}
	}
	// 如果提供了任务ID，获取命令和环境变量
	if req.TaskId > 0 {
		// 获取命令
		commands, err := s.getTaskCommands(req.TaskId)
		if err != nil {
			return nil, err
		}
		req.Command = commands

		// 获取环境变量
		envVars, err := s.getTaskEnvVars(req.TaskId)
		if err != nil {
			return nil, err
		}
		req.EnvVars = envVars
	}

	// 如果是定时触发，检查时间是否已到
	if req.TriggerType == "schedule" {
		// 将时间字符串转换为时间戳
		scheduleTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.Schedule, time.Local)
		if err != nil {
			return nil, fmt.Errorf("时间格式错误，请使用 YYYY-MM-DD HH:mm:ss 格式: %v", err)
		}

		scheduleTimestamp := scheduleTime.Unix()
		now := time.Now().Unix()

		if scheduleTimestamp > now {
			// 如果还没到时间，直接返回
			return &dto.ArgoCDDeployResponse{
				Name:      req.Name,
				Namespace: username,
				Status:    "pending",
				Message:   fmt.Sprintf("部署将在 %s 触发", req.Schedule),
			}, nil
		}
	}

	// 1. 生成Kubernetes部署YAML
	K8sDeployment := &K8sDeployment{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
	}

	// 设置元数据
	K8sDeployment.Metadata.Name = req.Name
	K8sDeployment.Metadata.Namespace = username // 使用用户名作为命名空间
	K8sDeployment.Metadata.Labels = req.Labels

	// 设置规格
	replicas := int32(req.Replicas)
	if replicas == 0 {
		replicas = 1
	}
	K8sDeployment.Spec.Replicas = replicas

	// 设置选择器
	if K8sDeployment.Metadata.Labels == nil {
		K8sDeployment.Metadata.Labels = make(map[string]string)
	}
	K8sDeployment.Metadata.Labels["app"] = req.Name
	K8sDeployment.Spec.Selector.MatchLabels = map[string]string{"app": req.Name}
	K8sDeployment.Spec.Template.Metadata.Labels = map[string]string{"app": req.Name}

	// 设置容器
	container := struct {
		Name    string              `yaml:"name"`
		Image   string              `yaml:"image"`
		Command []string            `yaml:"command,omitempty"`
		Args    []string            `yaml:"args,omitempty"`
		Ports   []dto.ContainerPort `yaml:"ports,omitempty"`
		Env     []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"env,omitempty"`
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu,omitempty"`
				Memory string `yaml:"memory,omitempty"`
			} `yaml:"requests,omitempty"`
			Limits struct {
				CPU    string `yaml:"cpu,omitempty"`
				Memory string `yaml:"memory,omitempty"`
			} `yaml:"limits,omitempty"`
		} `yaml:"resources,omitempty"`
	}{
		Name:    req.Name,
		Image:   req.Image,
		Command: req.Command,
		Args:    req.Args,
		Ports:   req.Ports,
	}

	// 设置环境变量
	if len(req.EnvVars) > 0 {
		for k, v := range req.EnvVars {
			container.Env = append(container.Env, struct {
				Name  string `yaml:"name"`
				Value string `yaml:"value"`
			}{Name: k, Value: v})
		}
	}

	// 设置资源限制
	if req.CPURequest != "" {
		container.Resources.Requests.CPU = req.CPURequest
	}
	if req.MemoryRequest != "" {
		container.Resources.Requests.Memory = req.MemoryRequest
	}
	if req.CPULimit != "" {
		container.Resources.Limits.CPU = req.CPULimit
	}
	if req.MemoryLimit != "" {
		container.Resources.Limits.Memory = req.MemoryLimit
	}

	K8sDeployment.Spec.Template.Spec.Containers = []struct {
		Name    string              `yaml:"name"`
		Image   string              `yaml:"image"`
		Command []string            `yaml:"command,omitempty"`
		Args    []string            `yaml:"args,omitempty"`
		Ports   []dto.ContainerPort `yaml:"ports,omitempty"`
		Env     []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"env,omitempty"`
		Resources struct {
			Requests struct {
				CPU    string `yaml:"cpu,omitempty"`
				Memory string `yaml:"memory,omitempty"`
			} `yaml:"requests,omitempty"`
			Limits struct {
				CPU    string `yaml:"cpu,omitempty"`
				Memory string `yaml:"memory,omitempty"`
			} `yaml:"limits,omitempty"`
		} `yaml:"resources,omitempty"`
	}{container}

	// 序列化Kubernetes部署配置
	yamlData, err := yaml.Marshal(K8sDeployment)
	if err != nil {
		return nil, fmt.Errorf("序列化YAML失败: %v", err)
	}

	// 检查 GitLab 配置
	if s.gitlabConfig == nil {
		fmt.Println("❌ 错误: GitLab配置未初始化")
		return nil, fmt.Errorf("GitLab配置未初始化")
	}
	if s.gitlabConfig.BaseURL == "" {
		fmt.Println("❌ 错误: GitLab BaseURL 未配置")
		return nil, fmt.Errorf("GitLab BaseURL 未配置")
	}

	// 提取项目ID
	projectID := extractProjectIDFromURL(req.GitRepo)
	if projectID == "" {
		fmt.Printf("❌ 错误: 无法从Git仓库URL解析项目ID: %s\n", req.GitRepo)
		return nil, fmt.Errorf("无法从Git仓库URL解析项目ID: %s", req.GitRepo)
	}

	fmt.Printf("✅ 项目ID解析成功: %s\n", projectID)

	// 🔥 关键修改：检查文件是否存在并动态设置操作类型
	yamlFileExists, err := s.checkFileExists(projectID, req.GitBranch, req.YamlPath)
	if err != nil {
		return nil, fmt.Errorf("检查YAML文件状态失败: %v", err)
	}

	// 检查用户目录是否存在
	userDirPath := filepath.ToSlash(filepath.Join("deployments", username, ".gitkeep"))
	dirExists, err := s.checkFileExists(projectID, req.GitBranch, userDirPath)
	if err != nil {
		fmt.Printf("⚠️ 检查目录存在性失败: %v, 将继续创建\n", err)
	}

	commitActions := []GitlabCommitAction{}

	// 1. 如果目录不存在，创建.gitkeep文件
	if !dirExists {
		commitActions = append(commitActions, GitlabCommitAction{
			Action:   "create",
			FilePath: userDirPath,
			Content:  "# This file ensures the directory exists",
		})
		fmt.Printf("✅ 将创建目录: %s\n", filepath.Dir(userDirPath))
	} else {
		fmt.Printf("✅ 目录已存在: %s\n", filepath.Dir(userDirPath))
	}

	// 2. 根据文件存在性动态设置操作类型
	actionType := "create"
	if yamlFileExists {
		actionType = "update"
		fmt.Printf("✅ YAML文件已存在，将执行更新操作\n")
	} else {
		fmt.Printf("✅ YAML文件不存在，将执行创建操作\n")
	}

	commitActions = append(commitActions, GitlabCommitAction{
		Action:   actionType,
		FilePath: req.YamlPath,
		Content:  string(yamlData),
	})

	commitPayload := GitlabCommitPayload{
		Branch:        req.GitBranch,
		CommitMessage: fmt.Sprintf("Deploy %s to %s", req.Name, username),
		Actions:       commitActions,
	}

	// 发送请求到Gitlab API
	commitURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits", s.gitlabConfig.BaseURL, projectID)
	payloadBytes, err := json.Marshal(commitPayload)
	if err != nil {
		return nil, fmt.Errorf("序列化提交数据失败: %v", err)
	}

	fmt.Printf("✅ 提交文件路径: %s\n", req.YamlPath)
	fmt.Printf("✅ 提交操作数量: %d\n", len(commitActions))

	httpReq, err := http.NewRequest("POST", commitURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("PRIVATE-TOKEN", s.gitlabConfig.Token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	fmt.Printf("GitLab API响应状态: %d\n", resp.StatusCode)
	fmt.Printf("GitLab API响应内容: %s\n", string(body))

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Gitlab API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应获取提交ID
	var commitResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &commitResponse); err != nil {
		return nil, fmt.Errorf("解析提交响应失败: %v, 响应体: %s", err, string(body))
	}

	fmt.Printf("✅ 提交创建成功，提交ID: %s\n", commitResponse.ID)
	if req.Tag != 1 {
		// 更新部署状态
		deployment.Status = "created"
		if err := s.db.Save(deployment).Error; err != nil {
			return nil, fmt.Errorf("更新部署状态失败: %v", err)
		}
	} else {
		// 查找现有部署记录
		var existingDeployment model.ArgoCDDeployment
		if err := s.db.Where("name = ? AND tenant_id = ?", req.Name, tenantId).First(&existingDeployment).Error; err != nil {
			return nil, fmt.Errorf("查找现有部署记录失败: %v", err)
		}
		// 更新状态
		existingDeployment.Status = "created"
		if err := s.db.Save(&existingDeployment).Error; err != nil {
			return nil, fmt.Errorf("更新部署状态失败: %v", err)
		}
	}
	return &dto.ArgoCDDeployResponse{
		Name:      req.Name,
		Namespace: username,
		Status:    "created",
		Message:   "Deployment created and pushed to Git repository",
		GitCommit: commitResponse.ID,
	}, nil
}

// checkFileExists 检查文件是否在GitLab仓库中已存在
func (s *ArgoCDService) checkFileExists(projectID, branch, filePath string) (bool, error) {
	// 构建查询文件内容的API URL
	checkURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s?ref=%s",
		s.gitlabConfig.BaseURL,
		projectID,
		url.PathEscape(filePath), // 对文件路径进行URL编码
		branch)

	httpReq, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return false, err
	}
	httpReq.Header.Set("PRIVATE-TOKEN", s.gitlabConfig.Token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 如果返回200，说明文件存在
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	// 如果返回404，说明文件不存在
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	// 其他状态码，返回错误
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("检查文件状态失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

// 定义DeploymentEvent结构
type DeploymentEvent struct {
	TenantId       int    `json:"tenant_id"`
	DeploymentName string `json:"deployment_name"`
}

// 添加定时检查函数
func (s *ArgoCDService) ProcessScheduledDeployments() error {
	var deployments []model.ArgoCDDeployment
	now := time.Now().Unix()
	fmt.Printf("进入定时查询")
	// 查找需要执行的定时部署
	if err := s.db.Where("trigger_type = ? AND status = ? AND schedule <= ?", "schedule", "pending", now).Find(&deployments).Error; err != nil {
		fmt.Printf("查询定时部署失败: %v", err)
		return fmt.Errorf("查询定时部署失败: %v", err)
	}

	for _, deployment := range deployments {
		// 获取用户名
		var user model.TenantUser
		if err := s.db.First(&user, deployment.TenantId).Error; err != nil {
			log.Printf("获取用户信息失败: %v\n", err)
			continue
		}

		// 构建请求
		req := &dto.ArgoCDDeployRequest{
			Name:        deployment.Name,
			GitRepo:     deployment.GitRepo,
			GitBranch:   deployment.GitBranch,
			Image:       deployment.Image,
			TriggerType: deployment.TriggerType,
			Schedule:    deployment.Schedule,
			YamlPath:    deployment.YamlPath,
			Tag:         1, // 设置 tag 为 1
		}

		// 执行部署
		if _, err := s.CreateArgoCDDeployment(context.Background(), user.Username, req); err != nil {
			log.Printf("执行定时部署失败 %s: %v\n", deployment.Name, err)
			continue
		}

	}
	return nil
}

// getArgoCDToken 获取ArgoCD认证token
func (s *ArgoCDService) getArgoCDToken() (string, error) {
	data := map[string]string{
		"username": s.argoCDConfig.Username,
		"password": s.argoCDConfig.Password,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// 创建一个忽略证书验证的 HTTP 客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", s.argoCDConfig.BaseURL+"/api/v1/session", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取token失败，状态码：%d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

// GetApplicationStatus 获取ArgoCD应用状态
func (s *ArgoCDService) GetApplicationStatus(namespace string) (map[string]interface{}, error) {
	token, err := s.getArgoCDToken()
	if err != nil {
		return nil, fmt.Errorf("获取ArgoCD token失败: %v", err)
	}

	appName := fmt.Sprintf("%s-app", namespace)
	url := fmt.Sprintf("%s/api/v1/applications/%s", s.argoCDConfig.BaseURL, appName)
	fmt.Printf("请求的完整URL是：%s\n", url)
	// 创建带超时控制的 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second, // 设置30秒超时
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			// 可选的额外传输层配置
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	fmt.Printf("请求的完整token是：%s\n", token)

	resp, err := client.Do(req)
	if err != nil {
		// 处理超时和其他网络错误
		if os.IsTimeout(err) {
			return nil, fmt.Errorf("请求超时: %v", err)
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取响应体获取更多错误信息
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取应用状态失败，状态码：%d, 响应：%s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取包含的资源信息
	return extractResources(result), nil
}

// extractResources 从完整响应中提取资源信息
func extractResources(fullResponse map[string]interface{}) map[string]interface{} {
	// 创建只包含资源信息的新map
	resourcesResponse := make(map[string]interface{})

	// 复制基础状态信息
	if status, ok := fullResponse["status"]; ok {
		if statusMap, ok := status.(map[string]interface{}); ok {
			resourcesStatus := make(map[string]interface{})

			// 只保留资源相关字段
			if resources, ok := statusMap["resources"]; ok {
				resourcesStatus["resources"] = resources
			}
			if health, ok := statusMap["health"]; ok {
				resourcesStatus["health"] = health
			}
			if sync, ok := statusMap["sync"]; ok {
				resourcesStatus["sync"] = sync
			}

			resourcesResponse["status"] = resourcesStatus
		}
	}

	// 保留元数据基本信息（可选）
	if metadata, ok := fullResponse["metadata"]; ok {
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			basicMetadata := make(map[string]interface{})
			if name, ok := metadataMap["name"]; ok {
				basicMetadata["name"] = name
			}
			if namespace, ok := metadataMap["namespace"]; ok {
				basicMetadata["namespace"] = namespace
			}
			resourcesResponse["metadata"] = basicMetadata
		}
	}

	return resourcesResponse
}

// 获取任务的命令列表
func (s *ArgoCDService) getTaskCommands(taskId int) ([]string, error) {
	var commands []struct {
		Command string
	}
	if err := s.db.Table("task_run_command").Select("command").Where("task_id = ?", taskId).Find(&commands).Error; err != nil {
		return nil, fmt.Errorf("获取任务命令失败: %v", err)
	}

	result := make([]string, len(commands))
	for i, cmd := range commands {
		result[i] = cmd.Command
	}
	return result, nil
}

// 获取任务的环境变量
func (s *ArgoCDService) getTaskEnvVars(taskId int) (map[string]string, error) {
	var envVars []struct {
		Key   string
		Value string
	}
	if err := s.db.Table("task_environment_var").Select("`key`, value").Where("task_id = ?", taskId).Find(&envVars).Error; err != nil {
		return nil, fmt.Errorf("获取任务环境变量失败: %v", err)
	}

	result := make(map[string]string)
	for _, env := range envVars {
		result[env.Key] = env.Value
	}
	return result, nil
}
