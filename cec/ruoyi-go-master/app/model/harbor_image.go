package model

import "time"

type BaseImage struct {
	Id          int    `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Tag         string `json:"tag" gorm:"size:50;not null"`
	Description string `json:"description"`
	Status      string `json:"status" gorm:"size:1;default:0"` // 0-正常 1-禁用
}

type UserImage struct {
	Id          int        `json:"id" gorm:"primaryKey"`
	TenantId    int        `json:"tenantId" gorm:"not null"`
	Name        string     `json:"name" gorm:"size:100;not null"`
	Tag         string     `json:"tag" gorm:"size:50;not null"`
	BaseImageId *int       `json:"baseImageId"`
	WorkDir     string     `json:"workDir" gorm:"size:255"`
	Cmd         string     `json:"cmd"`
	Status      string     `json:"status" gorm:"size:1;default:0"` // 0-正常 1-禁用
	BaseImage   BaseImage  `json:"baseImage" gorm:"foreignKey:BaseImageId"`
	Tenant      TenantUser `json:"tenant" gorm:"foreignKey:TenantId"`
}

// 在已有代码基础上添加

type ImageTaskStatus string

const (
	TaskStatusPending  ImageTaskStatus = "pending"
	TaskStatusRunning  ImageTaskStatus = "running"
	TaskStatusComplete ImageTaskStatus = "complete"
	TaskStatusFailed   ImageTaskStatus = "failed"
	TaskTypeLoad       string          = "load" // 新增加载镜像任务类型
)

type ImageTask struct {
	Id        int             `json:"id" gorm:"primaryKey"`
	TenantId  int             `json:"tenantId" gorm:"not null"`
	TaskType  string          `json:"taskType" gorm:"size:20;not null"` // push 或 pull
	ImageName string          `json:"imageName" gorm:"size:100;not null"`
	Tag       string          `json:"tag" gorm:"size:50;not null"`
	Status    ImageTaskStatus `json:"status" gorm:"size:20;not null"`
	Message   string          `json:"message" gorm:"size:500"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Username  string          `gorm:"-" json:"username"` // 新增字段
}

type ImageLoadTask struct {
	ImageTask
	FilePath string `json:"filePath" gorm:"size:255"` // 上传的tar文件路径
}

// 添加TableName方法，确保使用image_task表
func (ImageLoadTask) TableName() string {
	return "image_task"
}

type BuildFromArchiveReq struct {
	BaseProject     string `form:"baseProject" json:"baseProject" binding:"required"`
	BaseImageName   string `form:"baseImageName" json:"baseImageName" binding:"required"`
	BaseTag         string `form:"baseTag" json:"baseTag" binding:"required"`
	TargetImageName string `form:"targetImageName" json:"targetImageName" binding:"required"`
	TargetTag       string `form:"targetTag" json:"targetTag" binding:"required"`
}
