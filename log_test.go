package logger

import (
	"fmt"
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
	Debug(content)
	Info("Info line2")
	Errorf("Error")
	Flush()
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
				Stack("runtime error:%v", err)
			default: // 非运行时错误
				Stack("error:%v", err)
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
