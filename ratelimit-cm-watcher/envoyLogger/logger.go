package envoyLogger

import (
	"fmt"

	zapLogger "imesh.ai/ratelimit-cm-watcher/logger"
)

type Logger struct {
	Debug bool
}

func (logger Logger) Debugf(format string, args ...interface{}) {
	if logger.Debug {
		zapLogger.L.Debug(fmt.Sprintf("[DEBUG] "+format+"\n", args...))
	}
}

func (logger Logger) Infof(format string, args ...interface{}) {
	zapLogger.L.Info(fmt.Sprintf("[INFO]"+format+"\n", args...))
}

func (logger Logger) Warnf(format string, args ...interface{}) {
	zapLogger.L.Warn(fmt.Sprintf("[WARN] "+format+"\n", args...))
}

func (logger Logger) Errorf(format string, args ...interface{}) {
	zapLogger.L.Error(fmt.Sprintf("[ERROR]"+format+"\n", args...))
}
