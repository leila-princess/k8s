package middleware

import (
	"backend/app/service"
	"backend/config" // 添加这行导入
	"backend/framework/redis"
	"backend/framework/response"
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		if auth == "" {
			response.NewError().SetCode(401).SetMsg("未登录").Json(ctx)
			ctx.Abort()
			return
		}

		// 解析token
		parts := strings.SplitN(auth, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.NewError().SetCode(401).SetMsg("无效的认证头").Json(ctx)
			ctx.Abort()
			return
		}

		claims := &service.Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.NewError().SetCode(401).SetMsg("无效的token").Json(ctx)
			ctx.Abort()
			return
		}

		// 检查Redis中的token
		key := fmt.Sprintf("token:%d", claims.UserId)
		storedToken, err := redis.Get(key)
		if err != nil || storedToken != parts[1] {
			response.NewError().SetCode(401).SetMsg("token已失效").Json(ctx)
			ctx.Abort()
			return
		}

		// 将用户信息存入上下文
		ctx.Set("userId", claims.UserId)
		ctx.Set("username", claims.Username)
		ctx.Set("isAdmin", claims.IsAdmin)

		ctx.Next()
	}
}

// AdminRequired 管理员权限检查中间件
func AdminRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isAdmin, exists := ctx.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			response.NewError().SetCode(403).SetMsg("需要管理员权限").Json(ctx)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// package middleware

// import (
// 	"ruoyi-go/app/security"
// 	"ruoyi-go/app/token"
// 	"ruoyi-go/common/types/constant"
// 	"ruoyi-go/framework/response"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// // 认证中间件
// func AuthMiddleware() gin.HandlerFunc {

// 	return func(ctx *gin.Context) {

// 		authUser := security.GetAuthUser(ctx)
// 		if authUser == nil {
// 			response.NewError().SetCode(401).SetMsg("未登录").Json(ctx)
// 			ctx.Abort()
// 			return
// 		}

// 		// 判断token临期，小于20分钟刷新
// 		if authUser.ExpireTime.Time.Before(time.Now().Add(time.Minute * 20)) {
// 			token.RefreshToken(ctx, authUser.UserTokenResponse)
// 		}

// 		if authUser.Status != constant.NORMAL_STATUS {
// 			response.NewError().SetCode(601).SetMsg("用户被禁用").Json(ctx)
// 			ctx.Abort()
// 			return
// 		}

// 		ctx.Next()
// 	}
// }
