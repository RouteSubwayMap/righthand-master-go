package database

import (
	"strings"
	"testing"

	"righthand-go/internal/config"
)

// TestInitDB_Guard_RejectsNonTestDatabases 反例：测试进程连接非 *_test 库必须被 InitDB 拒绝（fail-fast），
// 不得建立任何连接——这是「自动测试禁止触碰开发/生产主库」强约束的机械保障（docs/ai/rules.md §9）。
func TestInitDB_Guard_RejectsNonTestDatabases(t *testing.T) {
	cases := []struct {
		name   string
		dbName string
	}{
		{"开发主库", "righthand"},
		{"其他业务库", "righthand_auth"},
		{"生产库名", "prod_righthand"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.DatabaseConfig{
				Host:      "localhost",
				Port:      3306,
				Username:  "root",
				Password:  "root",
				DBName:    tc.dbName,
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Asia/Shanghai",
			}
			err := InitDB(cfg)
			if err == nil {
				t.Fatalf("测试进程连接非测试库 %q 必须被拒绝", tc.dbName)
			}
			if !strings.Contains(err.Error(), "禁止连接非测试库") {
				t.Fatalf("错误信息应命中测试库强约束文案，实际: %v", err)
			}
		})
	}
}

// TestIsTestProcess_TrueInTestBinary 正例：go test 进程内 isTestProcess 必须为 true（guard 生效前提）
func TestIsTestProcess_TrueInTestBinary(t *testing.T) {
	if !isTestProcess() {
		t.Fatal("go test 进程内 isTestProcess 应返回 true")
	}
}

// TestTestDBName_SuffixConsistency 机械校验：约定的测试库名必须符合 *_test 规则，guard 才会放行
func TestTestDBName_SuffixConsistency(t *testing.T) {
	if !strings.HasSuffix(TestDBName, "_test") {
		t.Fatalf("TestDBName(%s) 必须以 _test 结尾，否则会被 guard 拦截", TestDBName)
	}
}
