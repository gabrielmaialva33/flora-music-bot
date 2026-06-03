package ubot

import (
	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) getP2PConfigs(GAorB []byte) (*P2PConfig, error) {
	dhConfigRaw, err := ctx.app.MessagesGetDhConfig(0, 256)
	if err != nil {
		return nil, err
	}
	dhConfig := dhConfigRaw.(*tg.MessagesDhConfigObj)
	return &P2PConfig{
		DhConfig:   dhConfig,
		IsOutgoing: GAorB == nil,
		GAorB:      GAorB,
		// Buffer 1: connectCall desiste após 10s; sem buffer, o send em
		// UpdatePhoneCall (handle_updates.go) travaria a goroutine ao chegar tarde.
		WaitData: make(chan error, 1),
	}, nil
}
