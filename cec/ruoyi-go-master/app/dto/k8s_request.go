package dto

type TestPodRequest struct {
	Token     string `json:"token" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

type TestPodResponse struct {
	PodName string `json:"podName"`
	Message string `json:"message"`
}

type CreateDeploymentRequest struct {
	Namespace     string            `json:"namespace" binding:"required"`
	Name          string            `json:"name" binding:"required"`
	Image         string            `json:"image" binding:"required"`
	Command       []string          `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	EnvVars       map[string]string `json:"envVars,omitempty"`
	CPURequest    string            `json:"cpuRequest,omitempty"`
	MemoryRequest string            `json:"memoryRequest,omitempty"`
	CPULimit      string            `json:"cpuLimit,omitempty"`
	MemoryLimit   string            `json:"memoryLimit,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Replicas      int32             `json:"replicas,omitempty"`
	Ports         []ContainerPort   `json:"ports,omitempty"`
}

type CreateUserDeploymentRequest struct {
	Name          string            `json:"name" binding:"required"`
	Image         string            `json:"image" binding:"required"`
	Command       []string          `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	EnvVars       map[string]string `json:"envVars,omitempty"`
	CPURequest    string            `json:"cpuRequest,omitempty"`
	MemoryRequest string            `json:"memoryRequest,omitempty"`
	CPULimit      string            `json:"cpuLimit,omitempty"`
	MemoryLimit   string            `json:"memoryLimit,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Replicas      int32             `json:"replicas,omitempty"`
	Ports         []ContainerPort   `json:"ports,omitempty"`
}

type ContainerPort struct {
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	ContainerPort int32  `json:"containerPort" binding:"required" yaml:"containerPort"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"` // TCP, UDP
}

type DeploymentResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// ArgoCDDeployRequest ArgoCD部署请求
type ArgoCDDeployRequest struct {
	TaskId        int               `json:"task_id,omitempty"` // 任务ID，用于获取command和env信息
	Name          string            `json:"name" binding:"required"`
	Image         string            `json:"image" binding:"required"` // harbor仓库镜像地址
	Command       []string          `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	EnvVars       map[string]string `json:"envVars,omitempty"`
	CPURequest    string            `json:"cpuRequest,omitempty"`
	MemoryRequest string            `json:"memoryRequest,omitempty"`
	CPULimit      string            `json:"cpuLimit,omitempty"`
	MemoryLimit   string            `json:"memoryLimit,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Replicas      int32             `json:"replicas,omitempty"`
	Ports         []ContainerPort   `json:"ports,omitempty"`
	GitRepo       string            `json:"gitRepo" binding:"required"`                          // Git仓库地址
	GitBranch     string            `json:"gitBranch" binding:"required"`                        // Git分支
	YamlPath      string            `json:"yamlPath"`                                            // YAML文件路径
	TriggerType   string            `json:"triggerType" binding:"required,oneof=schedule event"` // 触发类型
	Schedule      string            `json:"schedule,omitempty"`                                  // 定时触发的时间，格式：YYYY-MM-DD HH:mm:ss                                  // Cron表达式，当triggerType为schedule时必填
	Tag           int               `json:"tag"`                                                 // 添加 tag
}

// ArgoCDApplicationStatus ArgoCD应用状态
type ArgoCDApplicationStatus struct {
	Health struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"health"`
	Summary struct {
		Images []string `json:"images"`
	} `json:"summary"`
	Resources []struct {
		Group     string `json:"group"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Status    string `json:"status"`
		Health    struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"health"`
	} `json:"resources"`
}

// ArgoCDDeployResponse ArgoCD部署响应
type ArgoCDDeployResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Health    string `json:"health"`    // 添加健康状态
	GitCommit string `json:"gitCommit"` // Git提交ID
}
