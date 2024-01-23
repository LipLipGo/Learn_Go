package logger

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func example() {
	var l Logger
	// 这种风格需要提前留好占位符，或者拼接起来
	l.Info("用户的微信 id %d", 123)
}

// zap 风格

type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

func exampleV1() {
	var l LoggerV1
	//
	l.Info("这是一个新用户", Field{Key: "union_id", Value: 123})
}

type Field struct {
	Key   string
	Value any
}

type LoggerV2 interface {

	// 这种风格要求 args 必须是偶数，并且以 key1 value1,key2 value2 的形式传递,这种风格没有办法通过编译器来约束,用户想怎么传都可以

	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func exampleV2() {
	var l LoggerV2
	l.Info("这是一个新用户", "union_id", 123)
}
