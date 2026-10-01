// Package logger 提供统一日志初始化能力：控制台输出 + 按日期切割的文件日志。
// 零第三方依赖：滚动写入由 dailyWriter 实现（logs/app-YYYY-MM-DD.log），
// 保留策略按文件名日期清理（max_age 天数 / max_backups 文件数）。
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"righthand-go/internal/config"

	"github.com/gin-gonic/gin"
)

const (
	dateLayout      = "2006-01-02"
	logExt          = ".log"
	defaultBaseName = "app"
	defaultFilePath = "./logs/app.log"

	outputStdout = "stdout"
	outputFile   = "file"
	outputBoth   = "both"
)

// activeFile 全局文件写入器，供 Close 释放句柄
var activeFile *dailyWriter

// Init 依据配置初始化全局日志（标准库 log + gin 默认 writer）。
// output 取值：stdout / file / both（缺省与未知值回退 stdout）。
// file/both 模式下启动即创建当日文件并执行一次保留清理，路径不可用时返回错误，由调用方决定降级策略。
func Init(cfg *config.LogConfig) error {
	w, file, err := newWriter(cfg, os.Stdout)
	if err != nil {
		return err
	}
	log.SetOutput(w)
	gin.DefaultWriter = w
	gin.DefaultErrorWriter = w
	activeFile = file

	mode := normalizeOutput(cfg.Output)
	if file != nil {
		log.Printf("[LOGGER] 日志初始化完成 output=%s 文件=%s", mode, file.currentPath())
	} else {
		log.Printf("[LOGGER] 日志初始化完成 output=%s", mode)
	}
	return nil
}

// Close 关闭当前日志文件句柄（进程退出前调用）
func Close() error {
	if activeFile == nil {
		return nil
	}
	return activeFile.Close()
}

// newWriter 按 output 组装写入器；stdout 参数便于测试注入缓冲捕获控制台输出
func newWriter(cfg *config.LogConfig, stdout io.Writer) (io.Writer, *dailyWriter, error) {
	mode := normalizeOutput(cfg.Output)
	switch mode {
	case outputFile, outputBoth:
		fw := newDailyWriter(cfg.FilePath, cfg.MaxAge, cfg.MaxBackups)
		if err := fw.ensureToday(); err != nil {
			return nil, nil, err
		}
		if mode == outputFile {
			return fw, fw, nil
		}
		return io.MultiWriter(stdout, fw), fw, nil
	default:
		return stdout, nil, nil
	}
}

// normalizeOutput 归一化 output 配置（未知值回退 stdout）
func normalizeOutput(output string) string {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case outputFile:
		return outputFile
	case outputBoth:
		return outputBoth
	default:
		return outputStdout
	}
}

// dailyWriter 按天写入 <dir>/<base>-YYYY-MM-DD.log；跨日自动切换新文件
type dailyWriter struct {
	dir        string
	base       string
	maxAge     int
	maxBackups int
	now        func() time.Time

	mu      sync.Mutex
	file    *os.File
	curDate string
}

// newDailyWriter 从 file_path 解析目录与文件名（如 ./logs/app.log -> logs/app-YYYY-MM-DD.log）
func newDailyWriter(filePath string, maxAge, maxBackups int) *dailyWriter {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		filePath = defaultFilePath
	}
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if strings.TrimSpace(base) == "" {
		base = defaultBaseName
	}
	return &dailyWriter{
		dir:        dir,
		base:       base,
		maxAge:     maxAge,
		maxBackups: maxBackups,
		now:        time.Now,
	}
}

// Write 写入日志；跨日时先切换文件（读写作均在锁内，保证并发安全与整行写入）
func (w *dailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	date := w.now().Format(dateLayout)
	if w.file == nil || w.curDate != date {
		if err := w.rotate(date); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

// ensureToday 启动时创建当日文件并清理历史文件（尽早暴露路径/权限问题）
func (w *dailyWriter) ensureToday() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.rotate(w.now().Format(dateLayout))
}

// Close 关闭文件句柄（关闭后再次写入会自动重新打开）
func (w *dailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

// rotate 关闭旧句柄并打开指定日期的文件，随后执行保留清理
func (w *dailyWriter) rotate(date string) error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	f, err := os.OpenFile(w.filePath(date), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	w.file = f
	w.curDate = date
	w.cleanup(date)
	return nil
}

// cleanup 保留策略：当日文件必保留；超过 max_age 天的历史文件删除；
// 清理后文件总数超过 max_backups 时，从最旧开始删除（0 表示不启用该项限制）
func (w *dailyWriter) cleanup(currentDate string) {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}

	type datedFile struct {
		name string
		date time.Time
	}
	prefix := w.base + "-"
	files := make([]datedFile, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, logExt) {
			continue
		}
		d, perr := time.ParseInLocation(dateLayout, strings.TrimSuffix(strings.TrimPrefix(name, prefix), logExt), time.Local)
		if perr != nil {
			continue
		}
		files = append(files, datedFile{name: name, date: d})
	}
	// 按日期倒序：最新在前（当日文件必为首个）
	sort.Slice(files, func(i, j int) bool { return files[i].date.After(files[j].date) })

	cur, _ := time.ParseInLocation(dateLayout, currentDate, time.Local)
	kept := 0
	for _, f := range files {
		if f.name == w.fileName(currentDate) {
			kept++
			continue
		}
		if w.maxAge > 0 && cur.Sub(f.date) > time.Duration(w.maxAge)*24*time.Hour {
			_ = os.Remove(filepath.Join(w.dir, f.name))
			continue
		}
		if w.maxBackups > 0 && kept+1 > w.maxBackups {
			_ = os.Remove(filepath.Join(w.dir, f.name))
			continue
		}
		kept++
	}
}

// fileName 当日日志文件名（如 app-2026-09-23.log）
func (w *dailyWriter) fileName(date string) string {
	return fmt.Sprintf("%s-%s%s", w.base, date, logExt)
}

func (w *dailyWriter) filePath(date string) string {
	return filepath.Join(w.dir, w.fileName(date))
}

// currentPath 当前正在写入的文件路径（未打开时返回当日路径）
func (w *dailyWriter) currentPath() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	date := w.curDate
	if date == "" {
		date = w.now().Format(dateLayout)
	}
	return w.filePath(date)
}
