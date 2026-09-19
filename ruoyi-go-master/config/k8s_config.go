package config

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"time"
)

// K8s资源限制配置
var (
	// DefaultContainerCPULimit 默认容器CPU限制
	DefaultContainerCPULimit = resource.MustParse("500m")
	// DefaultContainerMemoryLimit 默认容器内存限制
	DefaultContainerMemoryLimit = resource.MustParse("512Mi")
	// DefaultContainerCPURequest 默认容器CPU请求
	DefaultContainerCPURequest = resource.MustParse("200m")
	// DefaultContainerMemoryRequest 默认容器内存请求
	DefaultContainerMemoryRequest = resource.MustParse("256Mi")
)

// JWT配置
var (
	// TokenExpireDuration JWT token过期时间
	TokenExpireDuration = time.Hour * 24 * 7
	// JWTSecret JWT密钥
	JWTSecret = "your-jwt-secret-key" // 建议从环境变量或配置文件中读取
)
