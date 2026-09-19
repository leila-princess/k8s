package main

import (
	"backend/app/router"
	"backend/app/service"
	"backend/config"
	"backend/framework/dal"
	"log"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	// 👇 新增
	_ "backend/docs" // 这行必须导入，不然 swagger 不会注册文档

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title RuoYi-Go API
// @version 1.0
// @description RuoYi-Go 后台管理系统 API 文档
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 39.96.159.232:3000
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := config.Data.Mysql.Username + ":" + config.Data.Mysql.Password + "@tcp(" + config.Data.Mysql.Host + ":" + strconv.Itoa(config.Data.Mysql.Port) + ")/" + config.Data.Mysql.Database + "?charset=" + config.Data.Mysql.Charset + "&parseTime=True&loc=Local"

	// 初始化数据访问层
	dal.InitDal(&dal.Config{
		GomrConfig: &dal.GomrConfig{
			Dialector: mysql.Open(dsn),
			Opts: &gorm.Config{
				SkipDefaultTransaction: true, // 跳过默认事务
				NamingStrategy: schema.NamingStrategy{
					SingularTable: true,
				},
				Logger: logger.New(log.Default(), logger.Config{
					// LogLevel: logger.Silent, // 不打印日志
					LogLevel:                  logger.Error, // 打印错误日志
					IgnoreRecordNotFoundError: true,
				}),
			},
			MaxOpenConns: config.Data.Mysql.MaxOpenConns,
			MaxIdleConns: config.Data.Mysql.MaxIdleConns,
		},
		RedisConfig: &dal.RedisConfig{
			Host:     config.Data.Redis.Host,
			Port:     config.Data.Redis.Port,
			Database: config.Data.Redis.Database,
			Password: config.Data.Redis.Password,
		},
	})

	// 初始化 ArgoCD 服务并启动调度器
    argoCDService := service.NewArgoCDService(
        dal.GetDB(),
        config.Data.Gitlab.BaseURL,
        config.Data.Gitlab.Token,
    )
    startScheduler(argoCDService)
	// 设置模式
	gin.SetMode(config.Data.Server.Mode)

	// 初始化gin
	server := gin.New()

	// 使用恢复中间件
	server.Use(gin.Recovery())

	// 设置文件资源目录
	server.Static(config.Data.Ruoyi.UploadPath, config.Data.Ruoyi.UploadPath)

	// 注册业务路由
	router.Register(server)

	// 👇 注册 Swagger 路由
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 启动服务
	server.Run(":" + strconv.Itoa(config.Data.Server.Port))
}

func startScheduler(argoCDService *service.ArgoCDService) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			if err := argoCDService.ProcessScheduledDeployments(); err != nil {
				log.Printf("处理定时部署失败: %v\n", err)
			}
		}
	}()
}
