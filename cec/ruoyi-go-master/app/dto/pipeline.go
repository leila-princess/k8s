package dto

type CreatePipelineRequest struct {
    Name            string `json:"name" binding:"required"`
    BaseImage       string `json:"baseImage" binding:"required"`
    CpuRequest      string `json:"cpuRequest" binding:"required"`
    MemoryRequest   string `json:"memoryRequest" binding:"required"`
    ServicePort     int    `json:"servicePort" binding:"required"`
    HealthCheckPath string `json:"healthCheckPath"`
    GitRepo         string `json:"gitRepo" binding:"required"`
    GitBranch       string `json:"gitBranch" binding:"required"`
    CronTrigger     string `json:"cronTrigger"`
}

type TriggerPipelineRequest struct {
    PipelineId int `json:"pipelineId" binding:"required"`
}