package logger

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
