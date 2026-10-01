package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"righthand-go/internal/config"
	"righthand-go/internal/database"

	"github.com/stretchr/testify/assert"
)

// TestMain 为 handler 包全部测试统一初始化数据库连接。
// 原因：本包多个测试走真实 handler + service（contact/task/project），
// 依赖全局 database.DB；若不在此处初始化，先执行的测试会因 DB 为 nil 直接 panic。
// 强约束：自动测试仅允许使用独立测试库 righthand_test（docs/ai/rules.md §9）。
func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败（需在 righthand-go 目录运行测试）: %v\n", err)
		os.Exit(1)
	}
	// 强约束：自动测试仅允许使用独立测试库（docs/ai/rules.md §9）
	cfg.Database.DBName = database.TestDBName
	if err := database.InitDB(&cfg.Database); err != nil {
		fmt.Fprintf(os.Stderr, "连接测试库 %s 失败，请先执行 go run ./cmd/init-test-db（见 docs/ai/rules.md §9）: %v\n", database.TestDBName, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// assertBizCode 断言统一响应契约：HTTP 固定 200，业务码在 body.code。
// 对齐前端拦截器与 project_handler_test.go 的断言方式（utils.ErrorWithCode 不改变 HTTP 状态码）。
func assertBizCode(t *testing.T, w *httptest.ResponseRecorder, wantCode int) {
	t.Helper()
	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应 body 失败: %v, body=%s", err, w.Body.String())
	}
	assert.Equal(t, wantCode, resp.Code)
}
