package model

import "time"

type TenantUser struct {
    Id         int       `gorm:"primaryKey;autoIncrement" json:"id" swaggerignore:"true"`
    Username   string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
    Password   string    `gorm:"size:128;not null" json:"password"`
    Status     string    `gorm:"size:1;default:0" json:"status"` // 0正常 1停用
    IsAdmin    string    `gorm:"size:1;default:0" json:"isAdmin"` // 0否 1是
    CreateTime time.Time `gorm:"autoCreateTime" json:"createTime"`
    UpdateTime time.Time `gorm:"autoUpdateTime" json:"createTime"`
}