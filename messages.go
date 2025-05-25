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

	// if rand.Intn(5) != 0 {
	// 	return
	// }
	prompt := fmt.Sprintf(`Given the following message, reply with a single emoji that best represents the emotion or tone of the message. Do not include any text besides the emoji. Message: "%s"`, m.Content)
	go analyzeAndReact(s, m, prompt)
}

func analyzeAndReact(s *discordgo.Session, m *discordgo.MessageCreate, prompt string) {
	// key := os.Getenv("OPENAI_API_KEY")
	// if key == "" {
	// 	log.Println("Missing OPENAI_API_KEY in environment")
	// 	return //"", fmt.Errorf("can't fetch mood, OpenAI key is missing. Blame the dev")
	// 	//log.Fatal("OPENAI_API_KEY not found")
	// }

	//prompt := fmt.Sprintf(`What is the emotion of this message? Respond with one word (e.g., happy, sad, angry, excited, confused, disappointed, etc). Message: %s"`, m.Content)

	// resp, err := openai.NewClient(key).CreateChatCompletion(
	// 	context.Background(),
	// 	openai.ChatCompletionRequest{
	// 		Model: openai.GPT3Dot5Turbo,
	// 		Messages: []openai.ChatCompletionMessage{
	// 			{Role: "user", Content: prompt},
	// 		},
	// 		MaxTokens: 5,
	// 	},
	// )
	// if err != nil {
	// 	log.Println("OpenAI call failed:", err)
	// 	return
	// }

	// reply := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))
	// //emoji := emotionToEmoji(emotion)
	// emoji := reply

	// if isProbabyEmoji(emoji) {
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

		//DEBUG
		log.Println("Detected reply:", emoji)
		if emoji == "" {
			log.Println("Blank reply")
			return
		}
	}()

	// err = s.MessageReactionAdd(m.ChannelID, m.ID, emoji)
	// if err != nil {
	// 	log.Println("Failed to add emoji reactions:", err)
	// }
}

// func emotionToEmoji(emotion string) string {
// 	switch emotion {
// 	case "angry", "mad", "annoyed", "furious":
// 		return "😠"
// 	case "happy", "joy", "joyful", "pleased", "delighted":
// 		return "😄"
// 	case "sad", "unhappy", "down", "depressed":
// 		return "😢"
// 	case "confused", "unsure":
// 		return "😕"
// 	case "excited", "thrilled":
// 		return "🤩"
// 	case "frustrated", "grumpy":
// 		return "💥"
// 	case "love", "loving":
// 		return "❤️"
// 	case "neutral", "okay", "fine":
// 		return "😐"
// 	default:
// 		return ""
// 	}
// }
