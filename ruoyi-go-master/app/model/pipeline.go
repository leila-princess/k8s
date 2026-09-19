package model

type Pipeline struct {
    Id              int    `json:"id" gorm:"primaryKey"`
    Name            string `json:"name" gorm:"size:100;not null"`
    BaseImage       string `json:"baseImage" gorm:"size:255;not null"` // 基础镜像
    CpuRequest      string `json:"cpuRequest" gorm:"size:50"`          // CPU请求
    MemoryRequest   string `json:"memoryRequest" gorm:"size:50"`      // 内存请求
    ServicePort     int    `json:"servicePort"`                         // 服务端口
    HealthCheckPath string `json:"healthCheckPath" gorm:"size:255"`    // 健康检查路径
    GitRepo         string `json:"gitRepo" gorm:"size:255"`           // Git仓库地址
    GitBranch       string `json:"gitBranch" gorm:"size:100"`         // Git分支
    CronTrigger     string `json:"cronTrigger" gorm:"size:100"`       // Cron触发器表达式
    Status          string `json:"status" gorm:"size:20"`             // 状态
}