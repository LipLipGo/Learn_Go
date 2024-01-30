package logger

import "go.uber.org/zap"

// 使用自定义的接口封装 zap

type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger(l *zap.Logger) LoggerV1 {
	return &ZapLogger{
		l: l,
	}
}

func (z *ZapLogger) Debug(msg string, args ...Field) {
	z.l.Debug(msg, z.toArgs(args)...)

}

func (z *ZapLogger) Info(msg string, args ...Field) {
	z.l.Info(msg, z.toArgs(args)...)

}

func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.l.Warn(msg, z.toArgs(args)...)

}

func (z *ZapLogger) Error(msg string, args ...Field) {
	z.l.Error(msg, z.toArgs(args)...)
}

// 将自定义的 Field 转换成 zap 的Field
func (z *ZapLogger) toArgs(args []Field) []zap.Field {

	res := make([]zap.Field, 0, len(args))

	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}