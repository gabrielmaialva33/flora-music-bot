package modules

import (
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
)

func MonitorRooms() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	sem := make(chan struct{}, 20)

	for range ticker.C {
		for chatID, room := range core.GetAllRooms() {

			sem <- struct{}{}

			go func(chatID int64, r *core.RoomState) {
				defer func() { <-sem }()

				if !r.IsActiveChat() {
					/*
						// TODO: TEST IT AND INCREASE SLEEP TIME
						time.Sleep(5 * time.Second)

						if !r.IsActiveChat() {
							core.DeleteRoom(chatID)
							return
						}
					*/
					return
				}

				if r.IsPaused() {
					return
				}

				r.Parse()
				statusMsg := r.StatusMsg()
				if statusMsg == nil {
					return
				}

				markup := core.GetPlayMarkup(r.EffectiveChatID(), r, false)
				opts := &telegram.SendOptions{
					ReplyMarkup: markup,
					Entities:    statusMsg.Message.Entities,
				}
				statusMsg.Edit(statusMsg.Text(), opts)
			}(chatID, room)
		}
	}
}
