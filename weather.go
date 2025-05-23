package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

var weatherCodeToMood = map[int]string{
	0:  "🌞 Feeling clear and energized!",
	1:  "🌤️ A few clouds — stay positive!",
	2:  "🌥️ Cloudy vibes — a thoughtful day.",
	3:  "☁️ Fully cloudy — maybe cozy up inside?",
	45: "🌫️ Foggy — slow and steady mood.",
	48: "🌫️ Foggy with frost — stay chill.",
	51: "🌦️ Light drizzle — mellow and relaxed.",
	53: "🌦️ Drizzle — a soft and dreamy mood.",
	55: "🌧️ Heavy drizzle — calm, maybe a bit sleepy.",
	61: "🌦️ Light rain — perfect for reflection.",
	63: "🌧️ Rain — chill and stay grounded.",
	65: "🌧️ Heavy rain — time for a deep mood.",
	71: "🌨️ Light snow — playful and fresh.",
	73: "🌨️ Snowfall — serene and quiet energy.",
	75: "🌨️ Heavy snow — peaceful and introspective.",
	80: "🌦️ Rain showers — energetic and lively!",
	81: "🌧️ Heavy showers — ride the chaos!",
	82: "🌧️ Violent rain showers — dramatic feels!",
	95: "⛈️ Thunderstorm — intense and passionate!",
	96: "⛈️ Thunderstorm with hail — wild mood!",
	99: "⛈️ Severe thunderstorm — electrifying energy!",
}

var weatherCodeDescriptions = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	56: "Light freezing drizzle",
	57: "Dense freezing drizzle",
	61: "Slight rain",
	63: "Moderate rain",
	65: "Heavy rain",
	66: "Light freezing rain",
	67: "Heavy freezing rain",
	71: "Slight snowfall",
	73: "Moderate snowfall",
	75: "Heavy snowfall",
	77: "Snow grains",
	80: "Slight rain showers",
	81: "Moderate rain showers",
	82: "Violent rain showers",
	85: "Slight snow showers",
	86: "Heavy snow showers",
	95: "Thunderstorm: Slight or moderate",
	96: "Thunderstorm with slight hail",
	99: "Thunderstorm with heavy hail",
}

type WeatherData struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Current              Current `json:"current"`
	Hourly               Hourly  `json:"hourly"`
	Daily                Daily   `json:"daily"`
}

type Current struct {
	Time        string  `json:"time"`
	WeatherCode float64 `json:"weather_code"`
}

type Hourly struct {
	Time          []string  `json:"time"`
	Temperature2m []float64 `json:"temperature_2m"`
}

type Daily struct {
	Time        []string  `json:"time"`
	WeatherCode []float64 `json:"weather_code"`
}

func getWeather() (*WeatherData, error) {
	url := "https://api.open-meteo.com/v1/forecast?latitude=-37.814&longitude=144.9633&daily=weather_code&hourly=temperature_2m&current=weather_code&timezone=auto&forecast_days=1"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching weather: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var data WeatherData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &data, nil
}

func handleMoodWeatherCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Step 1: Get weather
	weatherData, err := getWeather()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Failed to fetch weather data!",
			},
		})
		return
	}

	// Step 2: Get mood
	weatherCode := int(weatherData.Current.WeatherCode)
	mood, ok := weatherCodeToMood[weatherCode]
	if !ok {
		mood = "🤔 Mood unknown — but you are awesome anyway!"
	}

	// Step 3: Reply
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: mood,
		},
	})
}

func handleAIWeatherCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Immediate response
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🤖 Thinking about your weather mood...",
		},
	})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("🔥 Panic recovered in weather goroutine: %v\n", r)
				//s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
				//	Content: ptr("Something went wrong inside mood generator"),
				//})
			}
		}()

		//go func() {
		log.Println("Executing moodweather command")

		msg, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "✅ Gathering weather data",
		})
		if err != nil {
			log.Println("FollowupMessageCreate failed:", err)
			return
		}
		if msg == nil {
			log.Println("FollowupMessageCreate returned nil msg with no error!")
			return
		}
		log.Printf("✅ FollowupMessage created with ID: %s", msg.ID)

		//fetch weather
		weatherData, err := getWeather()
		if err != nil {
			log.Println("getWeather() failed:", err) //LOGGING
			s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
				Content: ptr("❌ Failed to fetch weather data!"),
			})
			return
		}

		weatherCode := int(weatherData.Current.WeatherCode)
		weatherDescription, ok := weatherCodeDescriptions[weatherCode]
		if !ok {
			s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
				Content: ptr("Unknown weatherCode."),
			})
			return
		}

		aireply, err := generateMoodFromWeather(weatherDescription)
		if err != nil {
			log.Println("generatedMoodFromWeather() failed:", err) //LOGGING
			s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
				Content: ptr("Unknown"),
			})
			return
		}

		s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
			Content: ptr(weatherDescription + " " + aireply),
		})
	}()

}

// ##TODO refactor into commandhandler
func weatherHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !botEnabled {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "🚫 Bot is currently disabled.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if !allowedChannels[i.ChannelID] {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⛔ You can't use this command in this channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	cmd := i.ApplicationCommandData()
	switch cmd.Name {
	case "moodbot":
		if len(cmd.Options) > 0 {
			switch cmd.Options[0].Name {
			case "weather":

				//if i.Type == discordgo.InteractionApplicationCommand {
				//	switch i.ApplicationCommandData().Name {
				//	case "weather":
				fmt.Println("Executing moodweather command") // Debug log
				//handleMoodWeatherCommand(s, i)
				handleAIWeatherCommand(s, i)
			}
		}
	}
}

// ##TODO add role system as variable input.
func generateMoodFromWeather(desc string) (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Println("Missing OPENAI_API_KEY in environment")
		return "", fmt.Errorf("can't fetch mood, OpenAI key is missing. Blame the dev")
		//log.Fatal("OPENAI_API_KEY not found")
	}
	client := openai.NewClient(key)
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a sarcastic assistant who gives short mood suggestions based on weather.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Weather: %s", desc),
			},
		},
	})
	if err != nil {
		log.Printf("❌ OpenAI API call failed: %v", err)
		return "", err
	}

	if len(resp.Choices) == 0 {
		log.Println("OpenAI returned no choices")
		return "", fmt.Errorf("no choices return from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}

func ptr(s string) *string {
	return &s
}
