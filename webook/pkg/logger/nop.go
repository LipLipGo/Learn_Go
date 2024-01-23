package logger

// 这个 Logger 是什么也不打印，比如在自动化测试的时候，什么也不打印

type NopLogger struct {
}

func NewNopLogger() LoggerV1 {
	return &NopLogger{}
}

func (n *NopLogger) Debug(msg string, args ...Field) {
}

func (n *NopLogger) Info(msg string, args ...Field) {
}

func (n *NopLogger) Warn(msg string, args ...Field) {

}

func (n *NopLogger) Error(msg string, args ...Field) {

}
