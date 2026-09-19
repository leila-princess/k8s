package dto

import "backend/app/model"

type TenantLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TenantLoginResponse struct {
	Token string `json:"token"`
}

type CreateTenantRequest struct {
	Username      string                 `json:"username" binding:"required"`
	Password      string                 `json:"password" binding:"required"`
	ResourceQuota model.K8sResourceQuota `json:"resourceQuota"`
}

// type UpdateTenantRequest struct {
// 	Id            int                     `json:"id" binding:"required"`
// 	Username      string                  `json:"username,omitempty"`
// 	Password      string                  `json:"password,omitempty"`
// 	Status        string                  `json:"status,omitempty"`
// 	ResourceQuota *model.K8sResourceQuota `json:"resourceQuota,omitempty"`
// }

type UpdateTenantRequest struct {
	Id       int    `json:"id,omitempty" swaggerignore:"true"`       // 可选，通过ID更新
	Username string `json:"username,omitempty" swaggerignore:"true"` // 可选，通过用户名更新
	Password string `json:"password,omitempty"`                      // 可选，更新密码
	Status   string `json:"status,omitempty"`                        // 可选，更新状态
}

type UpdatePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}
