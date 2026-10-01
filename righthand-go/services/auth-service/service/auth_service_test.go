package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/services/auth-service/dto"
)

func setupTestDB() {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "root",
		DBName:   database.TestDBName, // 强约束：自动测试仅允许使用独立测试库（docs/ai/rules.md §9）
		Charset:  "utf8mb4",
		ParseTime: true,
		Loc:      "Asia/Shanghai",
	}
	database.InitDB(cfg)
}

func TestAuthService_Login(t *testing.T) {
	// 跳过需要数据库的测试
	t.Skip("Skipping test that requires database connection")
	
	// 设置测试数据库
	setupTestDB()

	service := NewAuthService()

	// 测试登录
	req := &dto.LoginRequest{
		Username: "admin",
		Password: "admin123",
	}

	response, err := service.Login(req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.NotNil(t, response.User)
	assert.Equal(t, "admin", response.User.Username)
}

func TestAuthService_GetUserInfo(t *testing.T) {
	// 跳过需要数据库的测试
	t.Skip("Skipping test that requires database connection")
	
	setupTestDB()

	service := NewAuthService()

	// 测试获取用户信息（假设用户ID为1）
	userInfo, err := service.GetUserInfo(1)

	assert.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "admin", userInfo.Username)
}
