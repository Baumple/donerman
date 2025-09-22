package order

import (
	"fmt"
	"log"
	"time"

	"github.com/baumple/donerman/args"
	"github.com/baumple/donerman/doner"
	"github.com/bwmarrin/discordgo"
)

func sendOrderSummary(s *discordgo.Session, userOrdersMap map[*discordgo.User][]Order) {
	orderSummaryEmbeds := []*discordgo.MessageEmbed{}
	userIds := []string{}

	for user, orders := range userOrdersMap {
		totalOrderValue := 0.0
		title := user.Username + " hat bestellt:"

		fields := []*discordgo.MessageEmbedField{}

		for _, order := range orders {
			embedField := discordgo.MessageEmbedField{
				Name:  order.ItemName,
				Value: fmt.Sprintf("%.02f€", order.Price),
			}
			fields = append(fields, &embedField)

			totalOrderValue += order.Price
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

	_, err := s.ChannelMessageSendComplex(*args.DonerChannel, &discordgo.MessageSend{
		Content:         "# :rotating_light: Die Bestellungen sind nun da! :rotating_light:",
		Embeds:          orderSummaryEmbeds,
		AllowedMentions: &discordgo.MessageAllowedMentions{Roles: args.DonerRoles, Users: userIds},
	})

	if err != nil {
		log.Fatalln("Could not send order summary: " + err.Error())
	}

}

func sendOrderMessage(s *discordgo.Session, dm *doner.DonerMan, o *orderState, expiry time.Time) error {
	hour, minutes, _ := expiry.Clock()
	var msg = fmt.Sprintf(
		"Es wird heute bei %s%s bestellt."+
			"\nDrop mal bitte was du bestellen willst (Einfach der Name von dem Ding)"+
			"\nDu hast bis %d:%02d Zeit",
		dm.Name, dm.Emoji, hour, minutes,
	)
	_, err := s.ChannelMessageSendComplex(o.userChannel.ID, &discordgo.MessageSend{
		Content: msg,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Menü",
						Style: discordgo.LinkButton,
						Emoji: &discordgo.ComponentEmoji{Name: "📜"},
						URL:   dm.MenuURL,
					},
				},
			},
		},
	})
	return err
}
