package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if !botEnabled {
		return
	}

	if !allowedChannels[m.ChannelID] {
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

func analyzeIntentHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if !botEnabled {
		return
	}
	if !allowedChannels[m.ChannelID] {
		return
	}

	// if rand.Intn(5) != 0 {
	// 	return
	// }
	go analyzeAndReact(s, m)
}

func analyzeAndReact(s *discordgo.Session, m *discordgo.MessageCreate) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Println("Missing OPENAI_API_KEY in environment")
		return //"", fmt.Errorf("can't fetch mood, OpenAI key is missing. Blame the dev")
		//log.Fatal("OPENAI_API_KEY not found")
	}

	prompt := fmt.Sprintf(`What is the emotion of this message? Respond with one word (e.g., happy, sad, angry, excited, confused, disappointed, etc). Message: %s"`, m.Content)

	resp, err := openai.NewClient(key).CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: prompt},
			},
			MaxTokens: 5,
		},
	)
	if err != nil {
		log.Println("OpenAI call failed:", err)
		return
	}

	emotion := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))
	emoji := emotionToEmoji(emotion)

	log.Println("Detected emotion:", emotion)
	if emoji == "" {
		log.Println("Blank emoji")
		return
	}

	err = s.MessageReactionAdd(m.ChannelID, m.ID, emoji)
	if err != nil {
		log.Println("Failed to add emoji reactions:", err)
	}
}

func emotionToEmoji(emotion string) string {
	switch emotion {
	case "angry", "mad", "annoyed", "furious":
		return "😠"
	case "happy", "joy", "joyful", "pleased", "delighted":
		return "😄"
	case "sad", "unhappy", "down", "depressed":
		return "😢"
	case "confused", "unsure":
		return "😕"
	case "excited", "thrilled":
		return "🤩"
	case "frustrated", "grumpy":
		return "💥"
	case "love", "loving":
		return "❤️"
	case "neutral", "okay", "fine":
		return "😐"
	default:
		return ""
	}
}
