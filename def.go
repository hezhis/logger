package logger

type logData struct {
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
	traceColor = "\033[32m%s\033[0m\t"
	debugColor = "\033[32m%s\033[0m\t"
	infoColor  = "\033[32m%s\033[0m\t"
	warnColor  = "\033[35m%s\033[0m\t"
	errorColor = "\033[31m%s\033[0m\t"
	stackColor = "\033[31m%s\033[0m\t"
	fatalColor = "\033[31m%s\033[0m\t"
)
