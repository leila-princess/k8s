package model

type ArgoCDDeployment struct {
	Id          int    `gorm:"primaryKey;autoIncrement"`
	TenantId    int    `gorm:"not null"` // 关联租户ID
	Name        string `gorm:"not null"` // 部署名称
	GitRepo     string `gorm:"not null"` // Git仓库地址
	GitBranch   string `gorm:"not null"` // Git分支
	YamlPath    string `gorm:"not null"` // YAML文件路径
	Image       string `gorm:"not null"` // 镜像地址
	TriggerType string `gorm:"not null"` // 触发类型：schedule(定时)、event(事件)
	Schedule    string
	Health      string `gorm:"-"`        // ArgoCD健康状态
	Status      string `gorm:"not null"` // 部署状态
	LastRun     int64  // 上次运行时间
	CreatedAt   int64  `gorm:"autoCreateTime"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"`
}
