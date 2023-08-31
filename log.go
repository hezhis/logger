package logger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gzjjyz/trace"
	"github.com/petermattis/goid"
)

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
	traceColor = "\033[32m[Trace]\033[0m\t"
	debugColor = "\033[32m[Debug]\033[0m\t"
	infoColor  = "\033[32m[Info]\033[0m\t"
	warnColor  = "\033[35m[Warn]\033[0m\t"
	errorColor = "\033[31m[Error]\033[0m\t"
	stackColor = "\033[31m[Stack]\033[0m\t"
	fatalColor = "\033[31m[Fatal]\033[0m\t"
)

const (
	ConsoleLogStyle_Simple = "simple"
	ConsoleLogStyle_Detail = "detail"
)

var IsOutputScreen = true

var (
	writer          *FileLoggerWriter
	level           = TraceLevel
	logName         string //日志名称
	skip            = 2    //跳过等级
	logPath         string
	hasInit         bool
	initMu          sync.Mutex
	consoleLogStyle string = ConsoleLogStyle_Detail
	logPrefix              = ""
)

// GetLevel 获取日志级别
func GetLevel() int {
	return level
}

// SetLevel 设置日志级别
func SetLevel(l int) {
	if l > fatalLevel || l < TraceLevel {
		level = TraceLevel
	} else {
		level = l
	}
}

// SetLogPath 设置日志路径
func SetLogPath(path string) {
	logPath = path
}

// SetLogPath 设置日志路径
func SetLogPrefix(prefix string) {
	logPrefix = prefix
}

func SetConsoleLogStyle(style string) {
	if style != ConsoleLogStyle_Simple && style != ConsoleLogStyle_Detail {
		return
	}
	consoleLogStyle = style
}

func HasInit() bool {
	return hasInit
}

// InitLogger 日志模块初始化函数,程序启动时调用
func InitLogger(name string) {
	if HasInit() {
		return
	}

	initMu.Lock()
	defer initMu.Unlock()

	// maybe other thread is doing init too.
	if HasInit() {
		return
	}

	defer func() {
		hasInit = true
	}()

	//log文件夹不存在则先创建
	if logPath == "" {
		logPath = DefaultLogPath
	}

	logName = name

	writer = NewFileLoggerWriter(logPath, LogFileMaxSize, 5, OpenNewFileByByDateHour, 100000)
	go func() {
		err := writer.Loop()
		if err != nil {
			panic(err)
		}
	}()
	pID := os.Getpid()
	pIDStr := strconv.FormatInt(int64(pID), 10)
	Info("==========================================")
	Info("===log:%v,pid:%v==logPath:%s==", name, pIDStr, logPath)
	Info("==========================================")
}

type InfoStyle int

const (
	InfoStyle_Detail InfoStyle = 1
	InfoStyle_Simple InfoStyle = 2
)

func GetDetailInfo(style InfoStyle) string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(skip, pc)
	f := runtime.FuncForPC(pc[skip])
	if nil == f || len(pc) <= skip {
		return ""
	}
	file, line := f.FileLine(pc[skip])
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	funcName := f.Name()
	//for i := len(funcName) - 1; i > 0; i-- {
	//	if funcName[i] == '.' {
	//		funcName = funcName[i+1:]
	//		break
	//	}
	//}
	var traceId string
	if traceId, _ = trace.Ctx.GetCurGTrace(goid.Get()); traceId == "" {
		traceId = "UNKNOWN"
	}

	if style == InfoStyle_Detail {
		return fmt.Sprintf("\033[32m["+logName+"] %s [trace:%s] [%s:%d %s] %s\033[0m ", time.Now().Format("01-02 15:04:05.9999"), traceId, file, line, funcName, logPrefix)
	}

	if style == InfoStyle_Simple {
		return fmt.Sprintf("\033[32m["+logName+"] %s [trace:%s] [%s:%d] %s\033[0m ", time.Now().Format("01-02 15:04:05.9999"), traceId, file, line, logPrefix)
	}
	return ""
}

func Flush() {
	writer.Flush()
}

func doWrite(curLv int, colorInfo, format string, v ...interface{}) {
	if level > curLv {
		return
	}

	var builder strings.Builder
	builder.WriteString(colorInfo)
	builder.WriteString(GetDetailInfo(InfoStyle_Detail))
	content := fmt.Sprintf(format, v...)
	// protect disk
	if utf8.RuneCountInString(content) > 1000 {
		content = string([]rune(content)[:1000]) + "..."
	}
	builder.WriteString(content)

	if curLv >= stackLevel {
		buf := make([]byte, 4096)
		l := runtime.Stack(buf, true)
		builder.WriteString("\n")
		builder.WriteString(string(buf[:l]))
	}

	writer.Write(builder.String() + "\n")

	if curLv == fatalLevel {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		tf := time.Now()
		ioutil.WriteFile(fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic", dir, logName, tf.Month(), tf.Day(), tf.Hour(), tf.Minute(), tf.Second()), []byte(builder.String()), fileMode)

		panic(builder.String())
	}

	if IsOutputScreen {
		if consoleLogStyle == ConsoleLogStyle_Simple && curLv == TraceLevel {
			return
		}

		if consoleLogStyle == ConsoleLogStyle_Simple {
			var consoleStringBuilder strings.Builder
			consoleStringBuilder.WriteString(colorInfo)
			consoleStringBuilder.WriteString(GetDetailInfo(InfoStyle_Simple))
			consoleStringBuilder.WriteString(fmt.Sprintf(format, v...))
			fmt.Println(consoleStringBuilder.String())
			return
		}
		fmt.Println(builder.String())
	}
}

func Trace(format string, v ...interface{}) {
	doWrite(TraceLevel, traceColor, format, v...)
}

// Debug 调试类型日志
func Debug(format string, v ...interface{}) {
	doWrite(DebugLevel, debugColor, format, v...)
}

// Warn 警告类型日志
func Warn(format string, v ...interface{}) {
	doWrite(WarnLevel, warnColor, format, v...)
}

// Info 程序信息类型日志
func Info(format string, v ...interface{}) {
	doWrite(InfoLevel, infoColor, format, v...)
}

// Error 错误类型日志
func Errorf(format string, v ...interface{}) {
	doWrite(ErrorLevel, errorColor, format, v...)
}

// Fatalf 致命错误类型日志
func Fatalf(format string, v ...interface{}) {
	doWrite(fatalLevel, fatalColor, format, v...)
}

// Stack 堆栈debug日志
func Stack(format string, v ...interface{}) {
	doWrite(stackLevel, stackColor, format, v...)
}

// ErrorfNoCaller 错误类型日志 不包含调用信息
func ErrorfNoCaller(format string, v ...interface{}) {
	if level <= ErrorLevel {
		var traceId string
		if traceId, _ = trace.Ctx.GetCurGTrace(goid.Get()); traceId == "" {
			traceId = "UNKNOWN"
		}

		var builder strings.Builder
		builder.WriteString(errorColor)
		timeInfo := fmt.Sprintf("["+logName+"] %s [trace:%s]", time.Now().Format("01-02 15:04:05.9999"), traceId)
		builder.WriteString(timeInfo)
		builder.WriteString(fmt.Sprintf(format, v...))
		writer.Write(builder.String() + "\n")

		if IsOutputScreen {
			fmt.Println(builder.String())
		}
	}
}

func DebugIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 3
		Debug(format, v...)
		skip = 2
	}
}
func InfoIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 3
		Info(format, v...)
		skip = 2
	}
}
func WarnIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 3
		Warn(format, v...)
		skip = 2
	}
}
func ErrorIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 3
		Errorf(format, v...)
		skip = 2
	}
}
func FatalIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 3
		Fatalf(format, v...)
		skip = 2
	}
}
func StackIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 3
		Stack(format, v...)
		skip = 2
	}
}
