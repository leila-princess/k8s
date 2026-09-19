package model

import "time"

type K8sNamespace struct {
    Id            int              `gorm:"primaryKey;autoIncrement" json:"id" swaggerignore:"true"`
    Name          string           `gorm:"size:64;not null" json:"name"`
    UserId        int              `gorm:"not null" json:"userId" swaggerignore:"true"`
    Status        string           `gorm:"size:1;default:0" json:"status" swaggerignore:"true"` // 0正常 1停用
    ResourceQuota K8sResourceQuota `gorm:"foreignKey:NamespaceId" json:"resourceQuota"`
    ServiceAccount string          `gorm:"size:64" json:"serviceAccount,omitempty"` //存储sa的名
    CreateTime    time.Time        `gorm:"autoCreateTime" json:"createTime" swaggerignore:"true"`
    UpdateTime    time.Time        `gorm:"autoUpdateTime" json:"createTime" swaggerignore:"true"`
}