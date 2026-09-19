package service

import (
	"backend/app/dto"
	"backend/app/model"
	"context"
	"fmt"

	"github.com/bndr/gojenkins"
	"gorm.io/gorm"
)

type PipelineService struct{}

func (s *PipelineService) CreatePipeline(db *gorm.DB, req dto.CreatePipelineRequest) error {
	// 创建Jenkins客户端
	ctx := context.Background()
	jenkins := gojenkins.CreateJenkins(nil, "http://localhost:7000", "admin", "123")
	_, err := jenkins.Init(ctx)
	if err != nil {
		return fmt.Errorf("连接Jenkins失败：%v", err)
	}

	// 创建Jenkins Pipeline任务 - 使用Kubernetes Pod模板
	jobXml := fmt.Sprintf(`<?xml version='1.1' encoding='UTF-8'?>
    <flow-definition plugin="workflow-job@1316.vd2290d3341a_f">
        <description>Auto generated pipeline for %s</description>
        <keepDependencies>false</keepDependencies>
        <properties/>
        <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps@3783.v858d37070333">
            <script>
                pipeline {
    agent {
        kubernetes {
            label 'jenkins-agent'
            yaml '''
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    label: jenkins-agent
    name: jenkins-agent
spec:
  containers:
  - name: jnlp
    image: docker.m.daocloud.io/jenkins/inbound-agent:bookworm-jdk17
    imagePullPolicy: IfNotPresent
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
'''
        }
    }
    stages {
        stage('拉取代码') {
            steps {
                git branch: '%s', url: '%s'
            }
        }
        stage('构建') {
            steps {
                echo 'Building..'
            }
        }
        stage('测试') {
            steps {
                echo 'Testing..'
            }
        }
        stage('打包镜像并推送') {
            steps {
                script {
                    docker.build("%s:${BUILD_NUMBER}", "--build-arg BASE_IMAGE=%s .")
                    docker.withRegistry('http://39.96.159.232', 'harbor-credentials') {
                        docker.image("%s:${BUILD_NUMBER}").push()
                    }
                }
            }
        }
        stage('集群部署') {
            steps {
                echo 'Deploying..'
            }
        }
    }
}
            </script>
            <sandbox>true</sandbox>
        </definition>
        <triggers>
            <hudson.triggers.TimerTrigger>
                <spec>%s</spec>
            </hudson.triggers.TimerTrigger>
        </triggers>
    </flow-definition>`,
		req.Name, req.GitBranch, req.GitRepo, req.Name, req.BaseImage, req.Name, req.CronTrigger)

	_, err = jenkins.CreateJob(ctx, jobXml, req.Name)
	if err != nil {
		return fmt.Errorf("创建Jenkins任务失败：%v", err)
	}

	// 保存到数据库
	pipeline := &model.Pipeline{
		Name:            req.Name,
		BaseImage:       req.BaseImage,
		CpuRequest:      req.CpuRequest,
		MemoryRequest:   req.MemoryRequest,
		ServicePort:     req.ServicePort,
		HealthCheckPath: req.HealthCheckPath,
		GitRepo:         req.GitRepo,
		GitBranch:       req.GitBranch,
		CronTrigger:     req.CronTrigger,
		Status:          "0", // 0-正常 1-停用
	}

	return db.Create(pipeline).Error
}

func (s *PipelineService) TriggerPipeline(db *gorm.DB, req dto.TriggerPipelineRequest) error {
	var pipeline model.Pipeline
	if err := db.First(&pipeline, req.PipelineId).Error; err != nil {
		return fmt.Errorf("流水线不存在：%v", err)
	}

	// 触发Jenkins构建
	ctx := context.Background()
	jenkins := gojenkins.CreateJenkins(nil, "http://localhost:7000", "admin", "123")
	_, err := jenkins.Init(ctx)
	if err != nil {
		return fmt.Errorf("连接Jenkins失败：%v", err)
	}

	_, err = jenkins.BuildJob(ctx, pipeline.Name, nil)
	if err != nil {
		return fmt.Errorf("触发构建失败：%v", err)
	}

	return nil
}
