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
	name           string      // 日志名字
	level          int         // 日志等级
	bScreen        bool        // 是否打印屏幕
	path           string      // 目录
	prefix         string      // 标识
	maxFileSize    int64       // 文件大小
	perm           os.FileMode // 文件权限
	writer         *FileLoggerWriter
	goroutineTrace bool
}

type ILogger interface {
	LogWarn(format string, v ...interface{})
	LogInfo(format string, v ...interface{})
	LogError(format string, v ...interface{})
	LogFatal(format string, v ...interface{})
	LogDebug(format string, v ...interface{})
	LogStack(format string, v ...interface{})
	LogTrace(format string, v ...interface{})
}

var (
	instance          *logger
	initMu            sync.Mutex
	baseSkip          = 3  //跳过等级
	globalSkipPkgPath bool // 跳过包路径 方法
)

type CallInfoSt struct {
	File     string
	Line     int
	FuncName string
}

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

func GetLevel() int {
	return instance.level
}

// SetMaxSize 设置日志切割大小
func SetMaxSize(l int64) {
	if nil != instance {
		instance.maxFileSize = l
	}
}

// SetPerm 设置日志权限
func SetPerm(l os.FileMode) {
	if nil != instance {
		instance.perm = l
	}
}

func InitLogger(opts ...Option) ILogger {
	initMu.Lock()
	defer initMu.Unlock()

	if nil == instance {
		instance = &logger{
			goroutineTrace: true,
		}
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

	if instance.maxFileSize == 0 {
		instance.maxFileSize = LogFileMaxSize
	}

	if instance.perm == 0 {
		instance.perm = fileMode
	}

	if instance.writer == nil {
		instance.writer = NewFileLoggerWriter(instance.path, instance.maxFileSize, 5, OpenNewFileByByDateHour, 100000, instance.perm)

		go func() {
			err := instance.writer.Loop()
			if err != nil {
				panic(err)
			}
		}()
	}

	pID := os.Getpid()
	pIDStr := strconv.FormatInt(int64(pID), 10)
	LogInfo("===log:%v,pid:%v==logPath:%s==", instance.name, pIDStr, instance.path)

	return instance
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

func GetCallInfo(skip int) *CallInfoSt {
	pc, callFile, callLine, ok := runtime.Caller(skip)
	var callFuncName string
	if ok {
		// 拿到调用方法
		callFuncName = runtime.FuncForPC(pc).Name()
	}
	filePath, fileFunc := getPackageName(callFuncName)

	if globalSkipPkgPath {
		fileFunc = ""
		filePath = path.Base(filePath)
	}

	return &CallInfoSt{
		File:     path.Join(filePath, path.Base(callFile)),
		Line:     callLine,
		FuncName: fileFunc,
	}
}

func Flush() {
	instance.writer.Flush()
}

func buildTimeInfo() string {
	return time.Now().Format("01-02 15:04:05.9999")
}

func buildTraceInfo() string {
	traceId := "UNKNOWN"
	if id, _ := trace.Ctx.GetCurGTrace(goid.Get()); id != "" {
		traceId = id
	}
	return traceId
}

func buildContent(format string, v ...interface{}) string {
	content := fmt.Sprintf(format, v...)
	// protect disk
	if size := utf8.RuneCountInString(content); size > 15000 {
		content = string([]rune(content)[:15000]) + "..."
	}
	return content
}

func buildStackInfo() string {
	buf := make([]byte, 4096)
	l := runtime.Stack(buf, true)
	return string(buf[:l])
}

func buildCallInfo(call *CallInfoSt) string {
	if call == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d %s", call.File, call.Line, call.FuncName)
}

func buildRecord(curLv int, colorInfo, timeInfo, traceInfo, callerInfo, prefix, content string) string {
	var builder strings.Builder

	var header string
	if instance.goroutineTrace {
		header = fmt.Sprintf("%s %s [%s] [trace: %s] ", timeInfo, prefix, callerInfo, traceInfo)
	} else {
		header = fmt.Sprintf("%s %s [%s] ", timeInfo, prefix, callerInfo)
	}

	builder.WriteString(fmt.Sprintf(colorInfo, header))

	builder.WriteString(content)
	if curLv >= StackLevel {
		builder.WriteString("\n")
		builder.WriteString(buildStackInfo())
	}

	builder.WriteString("\n")
	return builder.String()
}

// LogTrace 跟踪类型日志
func LogTrace(format string, v ...interface{}) {
	instance.LogTrace(format, v...)
}

// LogDebug 调试类型日志
func LogDebug(format string, v ...interface{}) {
	instance.LogDebug(format, v...)
}

// LogWarn 警告类型日志
func LogWarn(format string, v ...interface{}) {
	instance.LogWarn(format, v...)
}

// LogInfo 程序信息类型日志
func LogInfo(format string, v ...interface{}) {
	instance.LogInfo(format, v...)
}

// LogError 错误类型日志
func LogError(format string, v ...interface{}) {
	instance.LogError(format, v...)
}

// LogStack 堆栈debug日志
func LogStack(format string, v ...interface{}) {
	instance.LogStack(format, v...)
}

// LogFatal 致命错误类型日志
func LogFatal(format string, v ...interface{}) {
	instance.LogFatal(format, v...)
}

// LogTraceWithRequester 跟踪类型日志
func LogTraceWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogTraceWithRequester(requester, format, v...)
}

// LogDebugWithRequester 调试类型日志
func LogDebugWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogDebugWithRequester(requester, format, v...)
}

// LogWarnWithRequester 警告类型日志
func LogWarnWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogWarnWithRequester(requester, format, v...)
}

// InfoWithRequester 程序信息类型日志
func LogInfoWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogInfoWithRequester(requester, format, v...)
}

func LogErrorWithRequesterAndCustomCallInfo(requester IRequester, callInfo *CallInfoSt, format string, v ...interface{}) {
	instance.LogErrorWithRequesterAndCustomCallInfo(requester, callInfo, format, v...)
}

// LogErrorWithRequester 错误类型日志
func LogErrorWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogErrorWithRequester(requester, format, v...)
}

// LogStackWithRequester 堆栈debug日志
func LogStackWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogStackWithRequester(requester, format, v...)
}

// LogFatalWithRequester 致命错误类型日志
func LogFatalWithRequester(requester IRequester, format string, v ...interface{}) {
	instance.LogFatalWithRequester(requester, format, v...)
}

// LogTraceWithRequester 跟踪类型日志
func (l *logger) LogTraceWithRequester(requester IRequester, format string, v ...interface{}) {
	if l.level > TraceLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format
	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	record := buildRecord(TraceLevel, traceColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

// LogDebugWithRequester 调试类型日志
func (l *logger) LogDebugWithRequester(requester IRequester, format string, v ...interface{}) {
	if l.level > DebugLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format
	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	record := buildRecord(DebugLevel, debugColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

// LogWarnWithRequester 警告类型日志
func (l *logger) LogWarnWithRequester(requester IRequester, format string, v ...interface{}) {
	if l.level > WarnLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format

	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	record := buildRecord(WarnLevel, warnColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

// InfoWithRequester 程序信息类型日志
func (l *logger) LogInfoWithRequester(requester IRequester, format string, v ...interface{}) {
	if l.level > InfoLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format

	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	record := buildRecord(InfoLevel, infoColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) LogErrorWithRequesterAndCustomCallInfo(requester IRequester, callInfo *CallInfoSt, format string, v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format

	record := buildRecord(ErrorLevel, errorColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

// LogErrorWithRequester 错误类型日志
func (l *logger) LogErrorWithRequester(requester IRequester, format string, v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format

	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	record := buildRecord(ErrorLevel, errorColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

// LogStackWithRequester 堆栈debug日志
func (l *logger) LogStackWithRequester(requester IRequester, format string, v ...interface{}) {
	if l.level > StackLevel {
		return
	}

	prefix := requester.GetLogPrefix()
	format = prefix + format

	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	record := buildRecord(StackLevel, stackColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

// LogFatalWithRequester 致命错误类型日志
func (l *logger) LogFatalWithRequester(requester IRequester, format string, v ...interface{}) {
	prefix := requester.GetLogPrefix()
	format = prefix + format

	callInfo := GetCallInfo(requester.GetLogCallStackSkip() + baseSkip)
	content := buildContent(format, v...)
	record := buildRecord(FatalLevel, fatalColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, content)
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
	os.Exit(1)
}

func (l *logger) LogWarn(format string, v ...interface{}) {
	if l.level > WarnLevel {
		return
	}

	callInfo := GetCallInfo(baseSkip)
	record := buildRecord(WarnLevel, warnColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) LogInfo(format string, v ...interface{}) {
	if l.level > InfoLevel {
		return
	}

	callInfo := GetCallInfo(baseSkip)
	record := buildRecord(InfoLevel, infoColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) LogError(format string, v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}

	callInfo := GetCallInfo(baseSkip)
	record := buildRecord(ErrorLevel, errorColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) LogFatal(format string, v ...interface{}) {
	callInfo := GetCallInfo(baseSkip)
	content := buildContent(format, v...)
	record := buildRecord(FatalLevel, fatalColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, content)
	l.writer.Write(record)

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	tf := time.Now()
	os.WriteFile(fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic", dir, l.name, tf.Month(), tf.Day(), tf.Hour(), tf.Minute(), tf.Second()), []byte(record), fileMode)

	if l.bScreen {
		fmt.Printf("%s", record)
	}

	os.Exit(1)
}

func (l *logger) LogDebug(format string, v ...interface{}) {
	if l.level > DebugLevel {
		return
	}

	callInfo := GetCallInfo(baseSkip)
	record := buildRecord(DebugLevel, debugColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) LogStack(format string, v ...interface{}) {
	if l.level > StackLevel {
		return
	}

	callInfo := GetCallInfo(baseSkip)
	record := buildRecord(StackLevel, stackColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)

	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) LogTrace(format string, v ...interface{}) {
	if l.level > TraceLevel {
		return
	}

	callInfo := GetCallInfo(baseSkip)
	record := buildRecord(TraceLevel, traceColor, buildTimeInfo(), buildTraceInfo(), buildCallInfo(callInfo), l.prefix, buildContent(format, v...))
	l.writer.Write(record)
	if l.bScreen {
		fmt.Printf("%s", record)
	}
}

func (l *logger) Flush() {
	l.writer.Flush()
}

func SetGlobalSkipFilePath() {
	globalSkipPkgPath = true
}
