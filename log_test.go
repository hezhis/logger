package logger

import (
	"fmt"
	"github.com/gzjjyz/srvlib/utils/signal"
	"log"
	"runtime"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	content := "你好吗"
	for i := 0; i < 1000; i++ {
		content += fmt.Sprintf("你好(%d)", i)
	}
	InitLogger(WithAppName("test"), WithLevel(InfoLevel), WithPrefix("pfId:1"), WithScreen(true))
	LogDebug(content)
	LogInfo("Info line2")
	LogError("Error")
	Flush()
}

func TestLogTime(t *testing.T) {
	content := "测试"
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	InitLogger(WithAppName("test"), WithLevel(InfoLevel), WithPrefix("pfId:1"), WithScreen(true))
	for range ticker.C {
		LogDebug(fmt.Sprintf("Debug %s(%d)", content, time.Now().Unix()))
		LogInfo(fmt.Sprintf("Info %s(%d)", content, time.Now().Unix()))
		LogError(fmt.Sprintf("Error %s(%d)", content, time.Now().Unix()))
	}

	<-signal.SignalChan()
}

func TestLog1(t *testing.T) {
	content := "你好吗"
	for i := 0; i < 1000; i++ {
		content += "你好吗"
	}
	InitLogger(WithAppName("test"), WithLevel(InfoLevel), WithPrefix("pfId:1"), WithScreen(true))
	go func() {
		// 延迟处理的函数
		defer func() {
			// 发生宕机时，获取panic传递的上下文并打印
			err := recover()
			if nil == err {
				return
			}
			switch err.(type) {
			case runtime.Error: // 运行时错误
				LogStack("runtime error:%v", err)
			default: // 非运行时错误
				LogStack("error:%v", err)
			}
		}()

		p()
	}()
	Flush()
	select {}
}

func p() {
	panic(fmt.Errorf("111"))
}

func TestTimeFormat(t *testing.T) {
	log.Println("gamesrv." + time.Now().Format("01-02-15.log"))
}
