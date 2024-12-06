package logger

import "go.uber.org/zap"

type ZapLogger struct {
	zapLogger *zap.Logger
}

func NewZapLogger(zapLogger *zap.Logger) Logger {
	return &ZapLogger{zapLogger: zapLogger}
}

func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.zapLogger.Debug(msg, fieldsToZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.zapLogger.Info(msg, fieldsToZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.zapLogger.Warn(msg, fieldsToZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.zapLogger.Error(msg, fieldsToZapFields(fields)...)
}

func fieldsToZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}