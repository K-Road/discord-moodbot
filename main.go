package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
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

var allowedChannels = map[string]bool{
	allowedChannelID: true,
}

var botEnabled = true
var messageCache *cache.Cache

func init() {
	messageCache = cache.New(10*time.Minute, 30*time.Minute)
}

func main() {
	//Load .env
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local testing
	}

	//Get token
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN not found")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord sessions: %v", err)
	}
	dg.ShouldReconnectOnError = true

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsMessageContent

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Print("✅ Bot is ready and connected to Discord.")
	})
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Disconnect) {
		log.Print("⚠️ Bot got disconnected from Discord!")
	})
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Connect) {
		log.Print("🔄 Bot reconnected to Discord.")
	})
	dg.AddHandler(func(s *discordgo.Session, evt *discordgo.Resumed) {
		log.Printf("🔁 Bot resumed session with Discord. Trace: %v", evt.Trace)
	})
	//DEBUG
	// dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 	log.Println("🔔 Raw message received:", m.Content)
	// })

	//dg.AddHandler(messageHandler)
	//dg.AddHandler(analyzeIntentHandler)
	dg.AddHandler(WrapWithCache(analyzeIntentHandler))

	dg.AddHandler(commandHandler)
	dg.AddHandler(weatherHandler)

	//Open discord Session
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	if err := registerCommands(dg); err != nil {
		log.Fatalf("Failed to register commands: %v", err)
	}

	defer dg.Close()

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello from discord bot!"))
		})
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	//fmt.Println("MoodBot is running. Press CTRL-C to exit.")
	select {}
}
