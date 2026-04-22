package core

import (
	"io"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
)

type TgLogger struct {
	gl  *gologging.Logger
	lvl telegram.LogLevel
}

func GetTgLogger(name string, lvl telegram.LogLevel) *TgLogger {
	l := &TgLogger{
		gl:  gologging.GetLogger(name),
		lvl: lvl,
	}
	l.SetLevel(lvl)
	l.SetOutput(config.LogWriter)
	return l
}

func (l *TgLogger) Debug(msg any, a ...any) {
	if l.lvl <= telegram.DebugLevel {
		l.gl.DebugF("%v %v", msg, a)
	}
}

func (l *TgLogger) Info(msg any, a ...any) {
	if l.lvl <= telegram.InfoLevel {
		l.gl.InfoF("%v %v", msg, a)
	}
}

func (l *TgLogger) Warn(msg any, a ...any) {
	if l.lvl <= telegram.WarnLevel {
		l.gl.WarnF("%v %v", msg, a)
	}
}

func (l *TgLogger) Error(msg any, a ...any) {
	if l.lvl <= telegram.ErrorLevel {
		l.gl.ErrorF("%v %v", msg, a)
	}
}

func (l *TgLogger) SetLevel(v telegram.LogLevel) {
	l.lvl = v
	switch v {
	case telegram.TraceLevel, telegram.DebugLevel:
		l.gl.SetLevel(gologging.DebugLevel)
	case telegram.InfoLevel:
		l.gl.SetLevel(gologging.InfoLevel)
	case telegram.WarnLevel:
		l.gl.SetLevel(gologging.WarnLevel)
	case telegram.ErrorLevel, telegram.PanicLevel:
		l.gl.SetLevel(gologging.ErrorLevel)
	case telegram.FatalLevel:
		l.gl.SetLevel(gologging.FatalLevel)
	default:
		l.gl.SetLevel(gologging.InfoLevel)
	}
}

func (l *TgLogger) GetLevel() telegram.LogLevel {
	return l.lvl
}

func (l *TgLogger) SetOutput(w any) {
	if ww, ok := w.(io.Writer); ok {
		l.gl.SetOutput(ww)
	}
}

func (l *TgLogger) GetOutput() any {
	return l.gl
}

func (l *TgLogger) SetTimestampFormat(s string) {}
