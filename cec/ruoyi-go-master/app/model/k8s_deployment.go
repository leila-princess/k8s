package model

import (
	"time"
)

type K8sDeployment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Namespace string    `gorm:"size:255;not null" json:"namespace"`
	Image     string    `gorm:"size:500" json:"image"`
	Replicas  int32     `json:"replicas"`
	Status    string    `gorm:"size:50" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
