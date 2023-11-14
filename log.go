package logger

import (
	"fmt"
	"os"
	"path"
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

type logger struct {
	name    string // 日志名字
	level   int    // 日志等级
	bScreen bool   // 是否打印屏幕
	path    string // 目录
	prefix  string // 标识
}

type ILogger interface {
	LogWarn(format string, v ...interface{})
	LogInfo(format string, v ...interface{})
	LogError(format string, v ...interface{})
	LogFatal(format string, v ...interface{})
	LogDebug(format string, v ...interface{})
	LogStack(format string, v ...interface{})
	LogTrace(format string, v ...interface{})
	LogErrorNoCaller(format string, v ...interface{})
}

var (
	instance *logger
	writer   *FileLoggerWriter
	initMu   sync.Mutex
	skip     = 3 //跳过等级
)

type IRequester interface {
	GetLogPrefix() string
	GetLogCallStackSkip() int
}

type DefaultLogRequester struct {
}

func (d *DefaultLogRequester) GetPrefix() string {
	return ""
}

// SetLevel 设置日志级别
func SetLevel(l int) {
	if l > FatalLevel || l < TraceLevel {
		return
	}
	if nil != instance {
		instance.level = l
	}
}

func InitLogger(opts ...Option) {
	initMu.Lock()
	defer initMu.Unlock()

	if nil == instance {
		instance = &logger{}
	}
	for _, opt := range opts {
		opt(instance)
	}

	//log文件夹不存在则先创建
	if instance.path == "" {
		dir := os.Getenv("TLOGDIR")
		if len(dir) > 0 {
			instance.path = dir
		} else {
			instance.path = DefaultLogPath
		}
	}

	if nil == writer {
		writer = NewFileLoggerWriter(instance.path, LogFileMaxSize, 5, OpenNewFileByByDateHour, 100000)

		go func() {
			err := writer.Loop()
			if err != nil {
				panic(err)
			}
		}()
	}

	pID := os.Getpid()
	pIDStr := strconv.FormatInt(int64(pID), 10)
	LogInfo("===log:%v,pid:%v==logPath:%s==", instance.name, pIDStr, instance.path)
}

func getPackageName(f string) (filePath string, fileFunc string) {
	slashIndex := strings.LastIndex(f, "/")
	filePath = f
	if slashIndex > 0 {
		idx := strings.Index(f[slashIndex:], ".") + slashIndex
		filePath = f[:idx]
		fileFunc = f[idx+1:]
		return
	}
	return
}

func GetDetailInfo(skip int) (file, funcName string, line int) {
	pc, callFile, callLine, ok := runtime.Caller(skip)
	var callFuncName string
	if ok {
		// 拿到调用方法
		callFuncName = runtime.FuncForPC(pc).Name()
	}
	filePath, fileFunc := getPackageName(callFuncName)
	return path.Join(filePath, path.Base(callFile)), fileFunc, callLine
}

func Flush() {
	writer.Flush()
}

func doWrite(callSkip int, curLv int, colorInfo, format string, v ...interface{}) {
	if curLv < instance.level {
		return
	}
	var builder strings.Builder

	file, funcName, line := GetDetailInfo(callSkip)

	traceId := "UNKNOWN"
	if id, _ := trace.Ctx.GetCurGTrace(goid.Get()); id != "" {
		traceId = id
	}
	detail := fmt.Sprintf("%s [%s:%d %s] ", time.Now().Format("01-02 15:04:05.9999"), file, line, funcName)
	detail = fmt.Sprintf(colorInfo, detail)
	builder.WriteString(detail)

	builder.WriteString(fmt.Sprintf("[trace:%s, prefix:%s] ", traceId, instance.prefix))

	content := fmt.Sprintf(format, v...)
	// protect disk
	if size := utf8.RuneCountInString(content); size > 10000 {
		content = "..." + string([]rune(content)[size-10000:])
	}
	builder.WriteString(content)

	if curLv >= StackLevel {
		buf := make([]byte, 4096)
		l := runtime.Stack(buf, true)
		builder.WriteString("\n")
		builder.WriteString(string(buf[:l]))
	}

	if curLv == FatalLevel {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		tf := time.Now()
		os.WriteFile(fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic", dir, instance.name, tf.Month(), tf.Day(), tf.Hour(), tf.Minute(), tf.Second()), []byte(builder.String()), fileMode)

		panic(builder.String())
	}
	builder.WriteString("\n")
	writer.Write(builder.String())

	if instance.bScreen {
		fmt.Printf("%s%s\n", detail, content)
	}
}

// LogTrace 跟踪类型日志
func LogTrace(format string, v ...interface{}) {
	doWrite(skip, TraceLevel, traceColor, format, v...)
}

// LogTraceWithRequester 跟踪类型日志
func LogTraceWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), TraceLevel, traceColor, format, v...)
}

// LogDebug 调试类型日志
func LogDebug(format string, v ...interface{}) {
	doWrite(skip, DebugLevel, debugColor, format, v...)
}

// LogDebugWithRequester 调试类型日志
func LogDebugWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), DebugLevel, debugColor, format, v...)
}

// LogWarn 警告类型日志
func LogWarn(format string, v ...interface{}) {
	doWrite(skip, WarnLevel, warnColor, format, v...)
}

// LogWarnWithRequester 警告类型日志
func LogWarnWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), WarnLevel, warnColor, format, v...)
}

// LogInfo 程序信息类型日志
func LogInfo(format string, v ...interface{}) {
	doWrite(skip, InfoLevel, infoColor, format, v...)
}

// InfoWithRequester 程序信息类型日志
func LogInfoWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), InfoLevel, infoColor, format, v...)
}

// LogError 错误类型日志
func LogError(format string, v ...interface{}) {
	doWrite(skip, ErrorLevel, errorColor, format, v...)
}

// LogErrorWithRequester 错误类型日志
func LogErrorWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), ErrorLevel, errorColor, format, v...)
}

// LogFatal 致命错误类型日志
func LogFatal(format string, v ...interface{}) {
	doWrite(skip, FatalLevel, fatalColor, format, v...)
}

// LogFatalWithRequester 致命错误类型日志
func LogFatalWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), FatalLevel, fatalColor, format, v...)
}

// LogStack 堆栈debug日志
func LogStack(format string, v ...interface{}) {
	doWrite(skip, StackLevel, stackColor, format, v...)
}

// LogStackWithRequester 堆栈debug日志
func LogStackWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(requester.GetLogCallStackSkip(), StackLevel, stackColor, format, v...)
}

// LogErrorNoCaller 错误类型日志 不包含调用信息
func LogErrorNoCaller(format string, v ...interface{}) {
	doWrite(0, ErrorLevel, errorColor, format, v...)
}

// LogErrorWithRequesterNoCaller 错误类型日志 不包含调用信息
func LogErrorWithRequesterNoCaller(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format
	doWrite(0, ErrorLevel, errorColor, format, v...)
}
