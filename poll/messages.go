package poll

import (
	"fmt"
	"log"
	"os"

	"github.com/baumple/donerman/args"
	"github.com/baumple/donerman/doner"
	"github.com/bwmarrin/discordgo"
)

func sendInvalidVoteMessage(s *discordgo.Session) {
	_, err := s.ChannelMessageSendComplex(
		*args.DonerChannel,
		&discordgo.MessageSend{
			Content: fmt.Sprintf("Niemand hat gevoted. Heute gibt es wohl kein %s D<@&:%s><@&:%s>",
				doner.GetRandomDonerName(), args.DonerRoles[0], args.DonerRoles[1],
			),
			AllowedMentions: &discordgo.MessageAllowedMentions{Roles: args.DonerRoles},
		},
	)
	if err != nil {
		log.Fatalln("Could not send invalid vote message: " + err.Error())
	}
	os.Exit(1)
}

func announcePollWinner(s *discordgo.Session, dm *doner.DonerMan) {
	var content = fmt.Sprintf(
		`# :rotating_light::rotating_light: Bestellwunschaufnahme!!! :rotating_light::rotating_light:
## Gewinner ist %s %s !!!!
<@&%s> <@&%s>

Weiteres wird per DM geklärt :saluting_face:!`,
		dm.Name,
		dm.Emoji,
		args.DonerRoles[0],
		args.DonerRoles[1],
	)
	_, err := s.ChannelMessageSendComplex(*args.DonerChannel, &discordgo.MessageSend{
		Content: content,
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: "Menu",
					Style: discordgo.LinkButton,
					URL:   dm.MenuURL,
					Emoji: &discordgo.ComponentEmoji{Name: "🏴‍☠️"},
				},
			},
		}},
	})

	if err != nil {
		log.Fatalln("Error while sending winner announcement: " + err.Error())
	}
}
