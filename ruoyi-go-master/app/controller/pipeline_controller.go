package controller

import (
	"backend/app/dto"
	"backend/app/service"
	"backend/framework/dal"
	"backend/framework/response"

	"github.com/gin-gonic/gin"
)

type PipelineController struct{}

// CreatePipeline 创建流水线
func (c *PipelineController) CreatePipeline(ctx *gin.Context) {
	var req dto.CreatePipelineRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	pipelineService := &service.PipelineService{}
	if err := pipelineService.CreatePipeline(dal.GetDB(), req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("创建成功").Json(ctx)
}

// TriggerPipeline 手动触发流水线
func (c *PipelineController) TriggerPipeline(ctx *gin.Context) {
	var req dto.TriggerPipelineRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	pipelineService := &service.PipelineService{}
	if err := pipelineService.TriggerPipeline(dal.GetDB(), req); err != nil {
		response.NewError().SetMsg(err.Error()).Json(ctx)
		return
	}

	response.NewSuccess().SetMsg("触发成功").Json(ctx)
}
