package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	content := "你好吗"
	for i := 0; i < 1000; i++ {
		content += "你好吗"
	}
	InitLogger(WithAppName("test"), WithLevel(InfoLevel), WithPrefix("pfId:1"), WithScreen(true))
	Debug(content)
	Info("Info\nline2")
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

func TestRead(t *testing.T) {
	f, err := os.Open(DefaultLogPath + "/" + "test.09-14-23.log")
	if nil != err {
		log.Println(err)
		return
	}
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		var isEOF bool
		if err != nil {
			if io.EOF != err {
				fmt.Println(err)
			} else {
				isEOF = true
			}
		}

		if len(line) > 0 {
			log.Println(line)
			data := LogData{}
			err := json.Unmarshal([]byte(line), &data)
			if nil != err {
				log.Println(err)
			}
		}

		if isEOF {
			break
		}
	}
}
