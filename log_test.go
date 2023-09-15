package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
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
			data := logData{}
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
