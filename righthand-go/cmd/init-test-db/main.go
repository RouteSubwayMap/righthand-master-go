package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// init-test-db 初始化/重建独立测试库 righthand_test（自动测试专用，强约束见 docs/ai/rules.md §9）。
//
// 用法（在 righthand-go 目录执行）：
//
//	go run ./cmd/init-test-db          # 仅初始化结构 + 系统基础数据（init-database.sql）
//	go run ./cmd/init-test-db -seed    # 额外导入演示业务数据（init-test-data.sql，首次执行）
//
// 原理：读取 SQL 脚本并将脚本内库目标 righthand 替换为 righthand_test 后整体执行，
// 避免脚本中硬编码的 USE righthand 把 DDL/DML 打到开发主库。
const (
	testDBName = "righthand_test"
	dsnFormat  = "root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&multiStatements=true&parseTime=true"
)

func main() {
	seed := flag.Bool("seed", false, "额外导入 init-test-data.sql（演示业务数据）")
	flag.Parse()

	if !strings.HasSuffix(testDBName, "_test") {
		fmt.Fprintf(os.Stderr, "拒绝执行：目标库名 %s 必须以 _test 结尾（防止误伤主库）\n", testDBName)
		os.Exit(1)
	}

	sqlDir, err := locateSQLDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	files := []string{"init-database.sql"}
	if *seed {
		files = append(files, "init-test-data.sql")
	}

	db, err := sql.Open("mysql", fmt.Sprintf(dsnFormat, testDBName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开连接失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	for _, name := range files {
		path := filepath.Join(sqlDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取 %s 失败: %v\n", path, err)
			os.Exit(1)
		}
		text := strings.TrimPrefix(string(content), "\ufeff")
		// 库目标重定向：防止脚本内 USE righthand 误伤开发主库（强约束相关，勿删）
		text = strings.ReplaceAll(text, "CREATE DATABASE IF NOT EXISTS righthand", "CREATE DATABASE IF NOT EXISTS "+testDBName)
		text = strings.ReplaceAll(text, "USE righthand;", "USE "+testDBName+";")

		if _, err := db.Exec(text); err != nil {
			fmt.Fprintf(os.Stderr, "执行 %s 失败: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("OK: 已导入 %s -> %s\n", name, testDBName)
	}
	fmt.Println("测试库初始化完成，可执行 go test（自动测试已强制使用 " + testDBName + "）")
}

// locateSQLDir 从当前工作目录定位包含 init-database.sql 的 sql 目录
func locateSQLDir() (string, error) {
	for _, c := range []string{"sql", filepath.Join("..", "sql")} {
		if _, err := os.Stat(filepath.Join(c, "init-database.sql")); err == nil {
			abs, err := filepath.Abs(c)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}
	return "", fmt.Errorf("未找到 sql/init-database.sql，请在 righthand-go 目录下执行 go run ./cmd/init-test-db")
}
