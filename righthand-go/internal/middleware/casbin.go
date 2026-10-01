package middleware

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"righthand-go/internal/utils"
)

var Enforcer *casbin.Enforcer

// InitCasbin 初始化Casbin
func InitCasbin(db *gorm.DB) error {
	// 创建Casbin适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return fmt.Errorf("failed to create casbin adapter: %v", err)
	}

	// 加载模型和策略
	Enforcer, err = casbin.NewEnforcer("configs/model.conf", adapter)
	if err != nil {
		return fmt.Errorf("failed to create casbin enforcer: %v", err)
	}

	// 加载策略
	if err := Enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load policy: %v", err)
	}

	return nil
}

// CasbinMiddleware Casbin权限中间件
func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("userID")
		if !exists {
			utils.Unauthorized(c, "未认证")
			c.Abort()
			return
		}

		// 获取请求路径和方法
		obj := c.Request.URL.Path
		act := c.Request.Method

		// 检查权限
		ok, err := Enforcer.Enforce(fmt.Sprintf("user:%d", userID), obj, act)
		if err != nil {
			utils.Error(c, "权限验证失败")
			c.Abort()
			return
		}

		if !ok {
			utils.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
