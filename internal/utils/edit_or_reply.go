package utils

import (
	"fmt"
	"runtime"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
)

func EOR(
	msg *telegram.NewMessage,
	text string,
	opts ...*telegram.SendOptions,
) (m *telegram.NewMessage, err error) {
	if msg == nil {
		gologging.Error("[EOR] nil msg at " + callerInfo(2))
		return nil, nil
	}

	m, err = msg.Edit(text, opts...)
	if err != nil {
		msg.Delete()
		m, err = msg.Respond(text, opts...)
	}

	if err != nil {
		gologging.Error(
			"[EOR] " + err.Error() +
				" | called from " + callerInfo(2),
		)
	}
	return m, err
}

func callerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", file, line)
}
