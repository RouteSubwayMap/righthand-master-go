package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
)

// ---------------------------------------------------------------------------
// 新增用户默认密码 / 姓名必输 集成测试（真库）
// 矩阵：tests/cases/user-create-default-password.md（7 场景与 7 个测试函数 1:1）
// ---------------------------------------------------------------------------

// newItUsername 生成带纳秒时间戳的测试用户名，避免与库中既有数据冲突
func newItUsername() string {
	return fmt.Sprintf("it_u%d", time.Now().UnixNano())
}

// adminBearer 生成管理员 JWT（userID=1，复用 config 已加载的 JWT 配置）
func adminBearer(t *testing.T) string {
	t.Helper()
	token, err := utils.GenerateToken(1, "admin", "管理员", true, "ADMIN")
	if err != nil {
		t.Fatalf("生成管理员 token 失败: %v", err)
	}
	return token
}

// doJSON 发送 JSON 请求（可选 Bearer token）
func doJSON(router *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// decodeResp 解析统一响应结构
func decodeResp(t *testing.T, w *httptest.ResponseRecorder) utils.Response {
	t.Helper()
	var resp utils.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	return resp
}

// cleanupUser 用例结束按用户名精确物理删除本用例创建的用户
func cleanupUser(t *testing.T, username string) {
	t.Helper()
	t.Cleanup(func() {
		database.GetDB().Unscoped().Where("username = ?", username).Delete(&models.SysUser{})
	})
}

// loginCode 使用用户名密码登录，返回业务 code（200=成功；401=未授权）
func loginCode(router *gin.Engine, username, password string) int {
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	w := doJSON(router, "POST", "/api/auth/login", body, "")
	var resp utils.Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Code
}

// createUserReq 发起创建用户请求
func createUserReq(router *gin.Engine, token, body string) *httptest.ResponseRecorder {
	return doJSON(router, "POST", "/api/system/users", body, token)
}

// Create_NoPassword_DefaultApplied：不携带 password 创建成功，默认密码 88888888 可登录
func TestCreateUser_NoPassword_DefaultApplied(t *testing.T) {
	router := setupTestRouter()
	username := newItUsername()
	cleanupUser(t, username)

	body := fmt.Sprintf(`{"username":%q,"realName":"集成测试用户"}`, username)
	w := createUserReq(router, adminBearer(t), body)
	resp := decodeResp(t, w)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	assert.Equal(t, 200, loginCode(router, username, "88888888"), "默认密码 88888888 应可登录")
}

// Create_EmptyPassword_DefaultApplied：显式空串密码与新契约等价，均走默认密码
func TestCreateUser_EmptyPassword_DefaultApplied(t *testing.T) {
	router := setupTestRouter()
	username := newItUsername()
	cleanupUser(t, username)

	body := fmt.Sprintf(`{"username":%q,"realName":"集成测试用户","password":""}`, username)
	w := createUserReq(router, adminBearer(t), body)
	resp := decodeResp(t, w)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, 200, loginCode(router, username, "88888888"))
}

// Create_ExplicitPassword_Ignored：强制默认策略下，传入的自定义密码被忽略
func TestCreateUser_ExplicitPassword_Ignored(t *testing.T) {
	router := setupTestRouter()
	username := newItUsername()
	cleanupUser(t, username)

	body := fmt.Sprintf(`{"username":%q,"realName":"集成测试用户","password":"Custom123"}`, username)
	w := createUserReq(router, adminBearer(t), body)
	resp := decodeResp(t, w)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, 200, loginCode(router, username, "88888888"), "默认密码应生效")
	assert.Equal(t, 401, loginCode(router, username, "Custom123"), "传入的自定义密码不应生效")
}

// Create_MissingRealName_Rejected：姓名缺失被 binding required 拦截且不写库
func TestCreateUser_MissingRealName_Rejected(t *testing.T) {
	router := setupTestRouter()
	username := newItUsername()

	body := fmt.Sprintf(`{"username":%q}`, username)
	w := createUserReq(router, adminBearer(t), body)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var count int64
	database.GetDB().Unscoped().Model(&models.SysUser{}).Where("username = ?", username).Count(&count)
	assert.Equal(t, int64(0), count, "姓名缺失不应写入任何用户")
}

// Create_MissingUsername_Rejected：用户名缺失被拦截
func TestCreateUser_MissingUsername_Rejected(t *testing.T) {
	router := setupTestRouter()

	w := createUserReq(router, adminBearer(t), `{"realName":"集成测试用户"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Create_DuplicateUsername_Rejected：同名二次创建被拒且库中仅一行
func TestCreateUser_DuplicateUsername_Rejected(t *testing.T) {
	router := setupTestRouter()
	username := newItUsername()
	cleanupUser(t, username)
	token := adminBearer(t)

	body := fmt.Sprintf(`{"username":%q,"realName":"集成测试用户"}`, username)
	first := createUserReq(router, token, body)
	assert.Equal(t, 200, decodeResp(t, first).Code)

	second := createUserReq(router, token, body)
	secondResp := decodeResp(t, second)
	assert.Equal(t, http.StatusOK, second.Code, "业务错误走 HTTP 200 + code=500")
	assert.Equal(t, 500, secondResp.Code)
	assert.Contains(t, secondResp.Message, "用户名已存在")

	var count int64
	database.GetDB().Unscoped().Model(&models.SysUser{}).Where("username = ?", username).Count(&count)
	assert.Equal(t, int64(1), count, "重复创建后库中应仅保留首次创建的一行")
}

// Create_StoredAsBcryptHash：密码以 bcrypt 加密入库且缺省状态为启用
func TestCreateUser_StoredAsBcryptHash(t *testing.T) {
	router := setupTestRouter()
	username := newItUsername()
	cleanupUser(t, username)

	body := fmt.Sprintf(`{"username":%q,"realName":"集成测试用户"}`, username)
	w := createUserReq(router, adminBearer(t), body)
	assert.Equal(t, 200, decodeResp(t, w).Code)

	var user models.SysUser
	err := database.GetDB().Unscoped().Where("username = ?", username).First(&user).Error
	assert.NoError(t, err)

	assert.NotEqual(t, "88888888", user.Password, "密码列不应为明文")
	assert.Len(t, user.Password, 60, "bcrypt hash 长度应为 60")
	assert.True(t, strings.HasPrefix(user.Password, "$2"), "bcrypt hash 应以 $2 开头")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("88888888")))
	assert.Equal(t, 1, user.Status, "缺省状态应为启用")
	assert.Equal(t, "集成测试用户", user.RealName)
}
