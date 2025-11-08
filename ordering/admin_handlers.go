package ordering

import (
	"log"
	"time"

	"github.com/baumple/donerman/ordering/orderstate"
	"github.com/baumple/donerman/utils"
	"github.com/bwmarrin/discordgo"
)

func advanceOrder(s *discordgo.Session,
	i *discordgo.InteractionCreate,
	state *orderstate.OrderState,
	_ map[string]*discordgo.ApplicationCommandInteractionDataOption,
) {
	state.OrderTimer.Reset(0)
	if utils.IsOberdirektor(i.Member) {
		utils.SendConfirmation(s, i)
	} else {
		utils.SendNotAllowed(s, i)
	}
}

func setOrderTime(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	state *orderstate.OrderState,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
) {
	if !utils.IsOberdirektor(i.Member) {
		utils.SendNotAllowed(s, i)
		return
	}
	var minutes time.Duration
	if m, ok := optionMap["minutes"]; ok {
		minutes = time.Duration(m.IntValue())
	}
	var seconds time.Duration
	if s, ok := optionMap["seconds"]; ok {
		seconds = time.Duration(s.IntValue())
	}

	duration := time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	timeEnd := time.Now().Add(duration)

	state.OrderTimer.Reset(duration)
	log.Println("Order now ends: " + timeEnd.Format("15:04"))

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Bestellung endet nun um: " + timeEnd.Format("15:04"),
		},
	})
	if err != nil {
		log.Println("Could not send confirmation message" + err.Error())
	}
}
