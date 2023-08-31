package logger

import (
	"testing"
)

func TestLog(t *testing.T) {
	content := "你好吗"
	for i := 0; i < 1000; i++ {
		content += "你好吗"
	}
	InitLogger("gamesrv")
	Debug(content)
	Flush()
}
