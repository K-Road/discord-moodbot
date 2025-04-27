package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func commandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "enable":
		//enable bot
		botEnabled = true
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Bot enabled!",
			},
		})
	case "disable":
		//disable bot
		botEnabled = false
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Bot disabled!",
			},
		})

	}
}

func registerCommands(dg *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "enable",
			Description: "Enable the bot",
		},
		{
			Name:        "disable",
			Description: "Disable the bot",
		},
		{
			Name:        "moodweather",
			Description: "Suggest a mood based on the current weather",
		},
	}

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("cannot create '%v' command: %w", cmd.Name, err)
		}
		log.Printf("registered: %s", cmd.Name)
	}

	return nil
}
