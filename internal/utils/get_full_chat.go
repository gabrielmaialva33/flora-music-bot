package utils

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetFullChannel(
	client *telegram.Client,
	chatID int64,
) (*telegram.ChannelFull, error) {
	peer, err := client.ResolvePeer(chatID)
	if err != nil {
		return nil, err
	}
	chPeer, ok := peer.(*telegram.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf(
			"chatID %d is not an InputPeerChannel, got %T",
			chatID,
			peer,
		)
	}

	fullChat, err := client.ChannelsGetFullChannel(&telegram.InputChannelObj{
		ChannelID:  chPeer.ChannelID,
		AccessHash: chPeer.AccessHash,
	})
	if err != nil {
		return nil, err
	}

	return fullChat.FullChat.(*telegram.ChannelFull), nil
}
