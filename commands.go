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
	//##TODO add allowed channel
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
	}
}

func registerCommands(dg *discordgo.Session) error {
	err := unregisterCommands(dg)
	if err != nil {
		return err
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "moodbot",
			Description: "Control Moodbot functionalities",
			//Type:        discordgo.ApplicationCommandType(discordgo.ApplicationCommandOptionSubCommandGroup),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "enable",
					Description: "Enable the bot",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "disable",
					Description: "Disable the bot",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "weather",
					Description: "Suggest a mood based on the current weather",
				},
			},
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

func unregisterCommands(dg *discordgo.Session) error {
	// Get all commands registered for the bot
	commands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("failed to get registered commands: %w", err)
	}

	// Delete all commands
	for _, cmd := range commands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
		if err != nil {
			log.Printf("Failed to delete command %s: %v", cmd.Name, err)
		} else {
			log.Printf("Deleted command: %s", cmd.Name)
		}
	}

	return nil
}
