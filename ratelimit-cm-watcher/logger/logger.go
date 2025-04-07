package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L *zap.Logger
)

func InitLogger() {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	zapConfig := zap.Config{
		Level:         zap.NewAtomicLevelAt(zap.ErrorLevel),
		Encoding:      "json",
		EncoderConfig: encoderCfg,
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	L = zap.Must(zapConfig.Build())
}

func Sync() {
	L.Sync()
}
