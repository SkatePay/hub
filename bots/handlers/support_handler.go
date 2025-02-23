package handlers

import (
	"fmt"
	"hub/core"
	"strings"

	"hub/bots"
)

type SupportHandler struct {
	Plugins []bots.HandlerPlugin // Handler-specific plugins
}

func (h *SupportHandler) HandleMessage(bot bots.Bot, message core.Message, senderPubKey string) {
	switch {
	case message.Content == "🙂":
		bot.PublishEncrypted(senderPubKey, "🙃")

	case strings.Contains(message.Content, "report"):
		reply := fmt.Sprintf("Could you elaborate on the problem you're encountering?")
		bot.PublishEncrypted(senderPubKey, reply)

	case message.Content == "I'm online.":
		welcomeMessage := "👋 Welcome to SkateConnect Support! Let us know if you need any assistance."
		bot.PublishEncrypted(senderPubKey, welcomeMessage)

		// 🔥 Trigger plugins directly for notifications
		h.TriggerPlugins(bot, message, senderPubKey)
	}
}

// TriggerPlugins allows handler plugins to react to specific events
func (h *SupportHandler) TriggerPlugins(bot bots.Bot, message core.Message, senderPubKey string) {
	for _, plugin := range h.Plugins {
		plugin.OnTrigger(bot, message, senderPubKey)
	}
}
