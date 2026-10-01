package utils

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"righthand-go/internal/config"
)

func TestGenerateToken(t *testing.T) {
	// 设置测试配置
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 3600,
			Issuer: "righthand-test",
		},
	}

	tests := []struct {
		name     string
		userID   int64
		username string
		realName string
		isAdmin  bool
		roles    string
		wantErr  bool
	}{
		{
			name:     "正常生成Token",
			userID:   1,
			username: "admin",
			realName: "管理员",
			isAdmin:  true,
			roles:    "admin,user",
			wantErr:  false,
		},
		{
			name:     "普通用户Token",
			userID:   2,
			username: "user1",
			realName: "普通用户",
			isAdmin:  false,
			roles:    "user",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.username, tt.realName, tt.isAdmin, tt.roles)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				
				// 验证生成的Token可以正确解析
				claims, parseErr := ParseToken(token)
				assert.NoError(t, parseErr)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.username, claims.Username)
				assert.Equal(t, tt.realName, claims.RealName)
				assert.Equal(t, tt.isAdmin, claims.IsAdmin)
				assert.Equal(t, tt.roles, claims.Roles)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	// 设置测试配置
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 3600,
			Issuer: "righthand-test",
		},
	}

	// 生成一个有效的Token
	validToken, _ := GenerateToken(1, "admin", "管理员", true, "admin")

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkUser bool
	}{
		{
			name:      "解析有效Token",
			token:     validToken,
			wantErr:   false,
			checkUser: true,
		},
		{
			name:      "解析无效Token",
			token:     "invalid.token.here",
			wantErr:   true,
			checkUser: false,
		},
		{
			name:      "解析空Token",
			token:     "",
			wantErr:   true,
			checkUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.token)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				
				if tt.checkUser {
					assert.Equal(t, int64(1), claims.UserID)
					assert.Equal(t, "admin", claims.Username)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	// 设置测试配置
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 3600,
			Issuer: "righthand-test",
		},
	}

	validToken, _ := GenerateToken(1, "admin", "管理员", true, "admin")

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "验证有效Token",
			token: validToken,
			want:  true,
		},
		{
			name:  "验证无效Token",
			token: "invalid.token",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateToken(tt.token)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// 设置测试配置 - Token 1秒后过期
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 1, // 1秒过期
			Issuer: "righthand-test",
		},
	}

	token, _ := GenerateToken(1, "admin", "管理员", true, "admin")
	
	// 立即验证应该成功
	assert.True(t, ValidateToken(token))
	
	// 等待2秒让Token过期
	time.Sleep(2 * time.Second)
	
	// 过期后验证应该失败
	assert.False(t, ValidateToken(token))
}

func TestGetUserIDFromToken(t *testing.T) {
	// 设置测试配置
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 3600,
			Issuer: "righthand-test",
		},
	}

	token, _ := GenerateToken(123, "admin", "管理员", true, "admin")

	userID, err := GetUserIDFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), userID)
}

func TestGetUsernameFromToken(t *testing.T) {
	// 设置测试配置
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 3600,
			Issuer: "righthand-test",
		},
	}

	token, _ := GenerateToken(1, "testuser", "测试用户", false, "user")

	username, err := GetUsernameFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
}
