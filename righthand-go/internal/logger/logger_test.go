package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"righthand-go/internal/config"
)

// 对应矩阵 tests/cases/app-logging.md

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 10, 0, 0, 0, time.Local)
}

// Logger_FileOutput_WritesDatedFile 正例：output=file 生成按日文件且内容一致
func TestLogger_FileOutput_WritesDatedFile(t *testing.T) {
	dir := t.TempDir()
	w := newDailyWriter(filepath.Join(dir, "app.log"), 30, 10)
	w.now = fixedNow(day(2026, time.September, 23))
	if err := w.ensureToday(); err != nil {
		t.Fatalf("初始化日志文件失败: %v", err)
	}
	defer w.Close()

	line := "启动日志\n"
	if _, err := w.Write([]byte(line)); err != nil {
		t.Fatalf("写入日志失败: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "app-2026-09-23.log"))
	if err != nil {
		t.Fatalf("期望生成 app-2026-09-23.log: %v", err)
	}
	if string(content) != line {
		t.Fatalf("文件内容期望 %q，实际 %q", line, string(content))
	}
}

// Logger_BothOutput_ConsoleAndFile 正例：output=both 同时写控制台与文件
func TestLogger_BothOutput_ConsoleAndFile(t *testing.T) {
	dir := t.TempDir()
	buf := &bytes.Buffer{}
	cfg := &config.LogConfig{Output: "both", FilePath: filepath.Join(dir, "app.log"), MaxAge: 30, MaxBackups: 10}
	w, file, err := newWriter(cfg, buf)
	if err != nil {
		t.Fatalf("初始化 both 输出失败: %v", err)
	}
	defer file.Close()
	file.now = fixedNow(day(2026, time.September, 23))

	line := "both输出-行\n"
	if _, err := w.Write([]byte(line)); err != nil {
		t.Fatalf("写入日志失败: %v", err)
	}
	if !strings.Contains(buf.String(), line) {
		t.Fatalf("stdout 缓冲期望包含 %q，实际 %q", line, buf.String())
	}
	content, err := os.ReadFile(filepath.Join(dir, "app-2026-09-23.log"))
	if err != nil {
		t.Fatalf("期望生成 app-2026-09-23.log: %v", err)
	}
	if !strings.Contains(string(content), line) {
		t.Fatalf("文件期望包含 %q，实际 %q", line, string(content))
	}
	if got := strings.Count(buf.String(), line); got != 1 {
		t.Fatalf("stdout 缓冲期望 1 行，实际 %d 行", got)
	}
}

// Logger_CrossDay_SwitchesToNewFile 边界：跨日写入自动切换新文件
func TestLogger_CrossDay_SwitchesToNewFile(t *testing.T) {
	dir := t.TempDir()
	w := newDailyWriter(filepath.Join(dir, "app.log"), 30, 10)
	cur := day(2026, time.September, 23)
	w.now = func() time.Time { return cur }
	if err := w.ensureToday(); err != nil {
		t.Fatalf("初始化日志文件失败: %v", err)
	}
	defer w.Close()

	if _, err := w.Write([]byte("day1\n")); err != nil {
		t.Fatalf("跨日前写入失败: %v", err)
	}
	cur = day(2026, time.September, 24)
	if _, err := w.Write([]byte("day2\n")); err != nil {
		t.Fatalf("跨日后写入失败: %v", err)
	}

	d1, err := os.ReadFile(filepath.Join(dir, "app-2026-09-23.log"))
	if err != nil {
		t.Fatalf("期望保留 app-2026-09-23.log: %v", err)
	}
	d2, err := os.ReadFile(filepath.Join(dir, "app-2026-09-24.log"))
	if err != nil {
		t.Fatalf("期望生成 app-2026-09-24.log: %v", err)
	}
	if string(d1) != "day1\n" {
		t.Fatalf("跨日后 09-23 文件不再追加，期望 %q，实际 %q", "day1\n", string(d1))
	}
	if string(d2) != "day2\n" {
		t.Fatalf("09-24 文件期望 %q，实际 %q", "day2\n", string(d2))
	}
}

// Logger_Retention_AgeThenBackupCount 边界：超龄删除 + 超量按最旧删除，当日文件必保留
func TestLogger_Retention_AgeThenBackupCount(t *testing.T) {
	dir := t.TempDir()
	oldDates := []string{"2026-08-23", "2026-08-25", "2026-08-26", "2026-08-27"}
	for _, d := range oldDates {
		if err := os.WriteFile(filepath.Join(dir, "app-"+d+".log"), []byte("old\n"), 0o644); err != nil {
			t.Fatalf("预置旧文件失败: %v", err)
		}
	}

	w := newDailyWriter(filepath.Join(dir, "app.log"), 30, 2)
	w.now = fixedNow(day(2026, time.September, 23))
	if err := w.ensureToday(); err != nil {
		t.Fatalf("初始化日志文件失败: %v", err)
	}
	defer w.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}
	got := make([]string, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.Name())
	}
	want := []string{"app-2026-08-27.log", "app-2026-09-23.log"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("保留策略期望剩余 %v（共 2 个等于 maxBackups），实际 %v", want, got)
	}
}

// Logger_StdoutOutput_NoFileCreated 反例：output=stdout 不创建任何文件
func TestLogger_StdoutOutput_NoFileCreated(t *testing.T) {
	dir := t.TempDir()
	buf := &bytes.Buffer{}
	cfg := &config.LogConfig{Output: "stdout", FilePath: filepath.Join(dir, "app.log")}
	w, file, err := newWriter(cfg, buf)
	if err != nil {
		t.Fatalf("初始化 stdout 输出失败: %v", err)
	}
	if file != nil {
		t.Fatalf("stdout 模式不应持有文件写入器")
	}

	line := "控制台行\n"
	if _, err := w.Write([]byte(line)); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if !strings.Contains(buf.String(), line) {
		t.Fatalf("stdout 缓冲期望包含 %q，实际 %q", line, buf.String())
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("stdout 模式不应创建文件，实际 %d 个", len(entries))
	}
}

// Logger_InvalidFilePath_InitFails 反例：路径不可用时初始化返回错误且不落文件
func TestLogger_InvalidFilePath_InitFails(t *testing.T) {
	dir := t.TempDir()
	fakeDir := filepath.Join(dir, "fake-dir")
	if err := os.WriteFile(fakeDir, []byte("occupied"), 0o644); err != nil {
		t.Fatalf("预置占位文件失败: %v", err)
	}

	cfg := &config.LogConfig{Output: "file", FilePath: filepath.Join(fakeDir, "logs", "app.log")}
	_, _, err := newWriter(cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatalf("父路径被普通文件占据时期望初始化失败")
	}
}

// Logger_ConcurrentWrites_AllLinesIntact 并发：50 个协程各写 1 行，行数与内容完整
func TestLogger_ConcurrentWrites_AllLinesIntact(t *testing.T) {
	dir := t.TempDir()
	w := newDailyWriter(filepath.Join(dir, "app.log"), 30, 10)
	w.now = fixedNow(day(2026, time.September, 23))
	if err := w.ensureToday(); err != nil {
		t.Fatalf("初始化日志文件失败: %v", err)
	}
	defer w.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if _, err := w.Write([]byte(fmt.Sprintf("line-%02d\n", n))); err != nil {
				t.Errorf("并发写入失败: %v", err)
			}
		}(i)
	}
	wg.Wait()

	content, err := os.ReadFile(filepath.Join(dir, "app-2026-09-23.log"))
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	if len(lines) != 50 {
		t.Fatalf("行数期望与写入次数一致 50，实际 %d", len(lines))
	}
	pattern := regexp.MustCompile(`^line-\d{2}$`)
	seen := map[string]bool{}
	for _, l := range lines {
		if !pattern.MatchString(l) {
			t.Fatalf("存在半行或交错内容: %q", l)
		}
		seen[l] = true
	}
	if len(seen) != 50 {
		t.Fatalf("期望 50 行各不相同，实际 %d 个唯一行", len(seen))
	}
}
