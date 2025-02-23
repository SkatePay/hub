package bots

type BotManager struct {
	Bots []Bot
}

// AddBot adds a bot to the manager
func (m *BotManager) AddBot(bot Bot) {
	m.Bots = append(m.Bots, bot)
}

// StartAll starts all bots concurrently
func (m *BotManager) StartAll() {
	for _, bot := range m.Bots {
		go bot.Start()
	}
}

// StopAll stops all bots
func (m *BotManager) StopAll() {
	for _, bot := range m.Bots {
		bot.Stop()
	}
}
