package plugins

import (
	"fmt"
	"hub/bots"
	"hub/core"
)

type ChannelNotifierPlugin struct {
	ChannelID string
}

// OnTrigger announces when specific message content is detected
func (c *ChannelNotifierPlugin) OnTrigger(bot bots.Bot, message core.Message, senderPubKey string) {
	if message.Content == "I'm online." {
		fmt.Printf("🔔 Channel %s: User %s has come online!\n", c.ChannelID, senderPubKey)
		core.Announce(
			c.ChannelID,
			senderPubKey,
			bot.GetSecretKey(),
			bot.GetPublicKey(),
		)
	}
}
