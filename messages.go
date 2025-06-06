package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/sashabaranov/go-openai"
)

type MessageHandlerFunc func(s *discordgo.Session, m *discordgo.MessageCreate)

func WrapWithCache(handler MessageHandlerFunc) MessageHandlerFunc {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		log.Println("DEBUG WrapWithCatch trigged. Message:", m.Content)
		if m.Author.Bot {
			return
		}
		if !botEnabled {
			return
		}
		if !allowedChannels[m.ChannelID] {
			return
		}
		normalized := strings.Join(strings.Fields(m.Content), " ")
		key := fmt.Sprintf("%s:%s", m.Author.ID, strings.ToLower(normalized))

		if _, found := messageCache.Get(key); found {
			log.Println("Message in cache:", key)
			return //Found in cache
		}
		log.Println("Processing message:", key)
		messageCache.Set(key, true, cache.DefaultExpiration)
		handler(s, m)
	}
}

// for keyword reactions
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

func analyzeMessageForEmoji(prompt string) (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Println("Missing OPENAI_API_KEY in environment")
		return "", fmt.Errorf("missinf OPENAI_API_KEY in env")
	}

	client := openai.NewClient(key)

	resp, err := client.CreateChatCompletion(
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
		return "", err
	}

	reply := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))
	if isProbabyEmoji(reply) {
		return reply, nil
	}
	return "", fmt.Errorf("not a valid emoji: %q", reply)
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

	//TODO Randomize react rate
	// if rand.Intn(5) != 0 {
	// 	return
	// }

	prompt := fmt.Sprintf(`Given the following message, reply with a single emoji that best represents the emotion or tone of the message. Do not include any text besides the emoji. Message: "%s"`, m.Content)
	go analyzeAndReact(s, m, prompt)
}

func analyzeAndReact(s *discordgo.Session, m *discordgo.MessageCreate, prompt string) {
	go func() {
		emoji, err := analyzeMessageForEmoji(prompt)
		if err != nil {
			log.Println("Invalid or blank emoji:", err)
			return
		}
		err = s.MessageReactionAdd(m.ChannelID, m.ID, emoji)
		if err != nil {
			log.Println("Failed to add reaction:", err)
		} else {
			log.Println("Invalid of blank emoji:", err)
		}

		//TODO check why blank
		log.Println("Detected reply:", emoji)
		if emoji == "" {
			log.Println("Blank reply")
			return
		}
	}()

}
