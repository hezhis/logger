package logger

import "os"

type Option func(log *logger)

func WithAppName(name string) Option {
	return func(log *logger) {
		log.name = name
	}
}

func WithPath(path string) Option {
	return func(log *logger) {
		log.path = path
	}
}

func WithLevel(level int) Option {
	return func(log *logger) {
		log.level = level
	}
}

func WithScreen(flag bool) Option {
	return func(log *logger) {
		log.bScreen = flag
	}
}

func WithPrefix(prefix string) Option {
	return func(log *logger) {
		log.prefix = prefix
	}
}

func WithPerm(perm os.FileMode) Option {
	return func(log *logger) {
		log.perm = perm
	}
}

func WithFileMaxSize(size int64) Option {
	return func(log *logger) {
		log.maxFileSize = size
	}
}

func WithoutGroutineTrace() Option {
	return func(log *logger) {
		log.groutineTrace = false
	}
}
