package logger

const (
	TraceLevel = iota // Trace级别
	DebugLevel        // Debug级别
	InfoLevel         // Info级别
	WarnLevel         // Warn级别
	ErrorLevel        // Error级别
	StackLevel        // stack级别
	FatalLevel        // Fatal级别
)

const (
	LogFileMaxSize   = 512
	fileMode         = 0777
	backupTimeFormat = "2006-01-02T15-04-05.000"
)

const (
	traceColor = "\033[32m[Trace] %s\033[0m"
	debugColor = "\033[32m[Debug] %s\033[0m"
	infoColor  = "\033[32m[Info] %s\033[0m"
	warnColor  = "\033[35m[Warn] %s\033[0m"
	errorColor = "\033[31m[Error] %s\033[0m"
	stackColor = "\033[31m[Stack] %s\033[0m"
	fatalColor = "\033[31m[Fatal] %s\033[0m"
)
