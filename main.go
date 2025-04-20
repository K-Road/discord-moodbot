package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var moodMap = map[string]string{
	"happy":   "😊",
	"joy":     "😄",
	"sad":     "😢",
	"tired":   "😴",
	"angry":   "😠",
	"love":    "❤️",
	"bored":   "🥱",
	"excited": "🤩",
}

const allowedChannelID = "1363353564109471935"

func main() {
	//Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//Get token
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN not found")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord sessions:", err)
		return
	}

	dg.AddHandler(messageHandler)
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}
	defer dg.Close()

	fmt.Println("MoodBot is running. Press CTRL-C to exit.")
	select {}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		fmt.Println("HERE")
		return
	}

	if m.ChannelID != allowedChannelID {
		fmt.Println("Reply to self")
		return
	}

	msg := strings.ToLower(m.Content)
	for keyword, emoji := range moodMap {
		if strings.Contains(msg, keyword) {
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, emoji)
			break
		}
	}

	if strings.HasPrefix(msg, "/moodbot") {
		s.ChannelMessageSend(m.ChannelID, "Hey! I'm MoodBot")
	}
}
