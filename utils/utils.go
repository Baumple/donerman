package utils

import (
	"log"
	"slices"

	"github.com/baumple/donerman/config"
	"github.com/bwmarrin/discordgo"
)

func SendConfirmation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Title:   "Jopp!",
			Content: "Ist gemacht",
		},
	})
	if err != nil {
		log.Println("Could not send confirmation: " + err.Error())
	}
}

func SendNotAllowed(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Title:   "Nope!",
			Content: "Nur der Oberdirektor darf das!!",
		},
	})
	if err != nil {
		log.Println("Could not send not allowed message: " + err.Error())
	}
}

func IsOberdirektor(m *discordgo.Member) bool {
	return m.User.ID == config.OberdirektorUser ||
		slices.Contains(m.Roles, config.OberdirektorRole)
}
