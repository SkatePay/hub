package bots

import (
	"hub/core"

	"github.com/nbd-wtf/go-nostr"
)

// Bot defines the interface for all bots
type Bot interface {
	Start()
	Stop()
	HandleEvent(event *nostr.Event)
	IsReady() bool
	PublishEncrypted(npubReceiver, message string)
	GetSecretKey() string
	GetPublicKey() string
}

// MessageHandler defines custom logic for handling messages
type MessageHandler interface {
	HandleMessage(bot Bot, message core.Message, senderPubKey string)
	TriggerPlugins(bot Bot, message core.Message, senderPubKey string)
}
