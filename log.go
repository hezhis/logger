package logger

import (
	"encoding/json"
	"fmt"
	"log"
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

var (
	instance *logger
	writer   *FileLoggerWriter
	initMu   sync.Mutex
	skip     = 3 //跳过等级
)

// SetLevel 设置日志级别
func SetLevel(l int) {
	if l > fatalLevel || l < TraceLevel {
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
	Info("===log:%v,pid:%v==logPath:%s==", instance.name, pIDStr, instance.path)
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

func GetDetailInfo() (file, funcName string, line int) {
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

func doWrite(curLv int, colorInfo, format string, v ...interface{}) {
	if instance.level > curLv {
		return
	}

	data := LogData{
		Level:     curLv,
		Timestamp: time.Now().Format("01-02 15:04:05.9999"),
		AppName:   instance.name,
		Prefix:    instance.prefix,
		color:     colorInfo,
	}

	if traceId, _ := trace.Ctx.GetCurGTrace(goid.Get()); traceId != "" {
		data.TraceId = traceId
	} else {
		data.TraceId = "UNKNOWN"
	}

	data.File, data.Func, data.Line = GetDetailInfo()

	content := fmt.Sprintf(format, v...)
	// protect disk
	if utf8.RuneCountInString(content) > 10000 {
		content = string([]rune(content)[:10000]) + "..."
	}
	data.Content = content

	if curLv >= stackLevel {
		buf := make([]byte, 4096)
		l := runtime.Stack(buf, true)
		data.Stack = string(buf[:l])
	}

	writer.Write(data)

	if curLv == fatalLevel {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		tf := time.Now()

		buf, err := json.Marshal(data)
		if nil != err {
			panic(err)
		} else {
			err := os.WriteFile(fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic", dir, instance.name, tf.Month(), tf.Day(),
				tf.Hour(), tf.Minute(), tf.Second()), buf, fileMode)
			if nil != err {
				log.Println(err)
			}
			panic(string(buf))
		}
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
	Errorf(format, v...)
}

func DebugIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 4
		Debug(format, v...)
		skip = 3
	}
}
func InfoIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 4
		Info(format, v...)
		skip = 3
	}
}
func WarnIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 4
		Warn(format, v...)
		skip = 3
	}
}
func ErrorIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 4
		Errorf(format, v...)
		skip = 3
	}
}
func FatalIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 4
		Fatalf(format, v...)
		skip = 3
	}
}
func StackIf(ok bool, format string, v ...interface{}) {
	if ok {
		skip = 4
		Stack(format, v...)
		skip = 3
	}
}
