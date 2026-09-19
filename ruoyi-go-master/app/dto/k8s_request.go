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

type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort" binding:"required"`
	Protocol      string `json:"protocol,omitempty"` // TCP, UDP
}

type DeploymentResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}
