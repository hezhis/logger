package logger

import (
	"fmt"
	"time"
)

type LogData struct {
	Level     int    `json:"level"`     // 日志等级
	Timestamp string `json:"timestamp"` // 时间
	AppName   string `json:"app_name"`  // 应用名字
	Content   string `json:"content"`   // 内容
	TraceId   string `json:"trace_id"`  // 链路id
	File      string `json:"file"`      // 文件名
	Line      int    `json:"line"`      // 行数
	Func      string `json:"func"`      // 函数名
	Prefix    string `json:"prefix"`    // 标识
	Stack     string `json:"stack"`     // 堆栈
	color     string
}

func (d *LogData) String() string {
	return fmt.Sprintf(" %s [%s:%d %s]", time.Now().Format("01-02 15:04:05.9999"), d.File, d.Line, d.Func)
}

const (
	TraceLevel = iota // Trace级别
	DebugLevel        // Debug级别
	InfoLevel         // Info级别
	WarnLevel         // Warn级别
	ErrorLevel        // Error级别
	stackLevel        // stack级别
	fatalLevel        // Fatal级别
)

const (
	LogFileMaxSize = 1024 * 1024 * 1024 * 10
	fileMode       = 0777
)

const (
	traceColor = "\033[32m[Trace] %s\033[0m"
	debugColor = "\033[32m[Debug] %s\033[0m"
	infoColor  = "\033[32m[Info] %s\033[0m"
	warnColor  = "\033[35m[Warn] %s\033[0m"
	errorColor = "\033[31m[Error] %s\033[0m"
	stackColor = "\033[31m[Stack] %s\033[0m"
	fatalColor = "\033[31m[Fatal] %s\033[0m"
)
