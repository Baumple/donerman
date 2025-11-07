package order

import (
	"fmt"
	"log"

	"github.com/baumple/donerman/config"
	"github.com/bwmarrin/discordgo"
)

func sendOrderSummary(
	s *discordgo.Session,
	userOrdersMap map[string][]Order,
	users map[string]*discordgo.User,
) {
	orderSummaryEmbeds := []*discordgo.MessageEmbed{}
	userIds := []string{}

	for userID, orders := range userOrdersMap {
		totalOrderValue := 0.0
		user := users[userID]

		title := user.Username + " hat bestellt:"

		fields := []*discordgo.MessageEmbedField{}
		for _, order := range orders {
			embedField := discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%d x %s", order.Amount, order.ItemName),
				Value: fmt.Sprintf("%d x %.02f€ = %0.2f (%s)",
					order.Amount,
					order.PricePerPiece,
					order.TotalPrice(),
					order.PaymentMethod.String(),
				),
			}
			fields = append(fields, &embedField)

			totalOrderValue += order.TotalPrice()
		}

		embed := discordgo.MessageEmbed{
			Type:   discordgo.EmbedTypeRich,
			Title:  title,
			Fields: fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Gesamt: %.02f€", totalOrderValue),
			},
			Color: 15418782,
		}
		orderSummaryEmbeds = append(orderSummaryEmbeds, &embed)
		userIds = append(userIds, user.ID)
	}

	_, err := s.ChannelMessageSendComplex(*config.DonerChannel, &discordgo.MessageSend{
		Content: "# :rotating_light: Die Bestellungen sind nun da! :rotating_light: @here",
		Embeds:  orderSummaryEmbeds,
	})

	if err != nil {
		log.Fatalln("Could not send order summary: " + err.Error())
	}

}
