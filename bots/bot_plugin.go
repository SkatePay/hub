package bots

import (
	"hub/core"

	"github.com/nbd-wtf/go-nostr"
)

// BotPlugin triggers globally on every event
type BotPlugin interface {
	OnEvent(bot Bot, event *nostr.Event)
}

// HandlerPlugin triggers when specific conditions are met (custom message handlers)
type HandlerPlugin interface {
	OnTrigger(bot Bot, message core.Message, senderPubKey string)
}
