package core

import "github.com/nbd-wtf/go-nostr"

type ContentStructure struct {
	Content string `json:"content"`
	Kind    string `json:"kind"`
}

type DefaultProvider struct {
	Relay      *nostr.Relay
	ChannelId  string
	PublicKey  string
	PrivateKey string
}

type Message struct {
	Content string `json:"content"`
	Kind    string `json:"kind"`
}

type RelayProvider interface {
	GetRelay() *nostr.Relay
	GetChannelId() string
	GetPrivateKey() string
	GetPublicKey() string
}
