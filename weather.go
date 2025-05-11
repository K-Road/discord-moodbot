package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
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
	// check if enabled
	// if !botEnabled {
	// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
	// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
	// 		Data: &discordgo.InteractionResponseData{
	// 			Content: "🚫 Bot is currently disabled.",
	// 			Flags:   discordgo.MessageFlagsEphemeral,
	// 		},
	// 	})
	// 	return
	// }

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

	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "weather":
			fmt.Println("Executing moodweather command") // Debug log
			handleMoodWeatherCommand(s, i)
		}
	}
}
