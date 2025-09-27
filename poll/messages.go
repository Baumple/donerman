package poll

import (
	"fmt"
	"log"
	"os"
	"time"

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

func announcePollWinner(s *discordgo.Session, dm *doner.DonerMan, expiry time.Time) {
	var content = fmt.Sprintf(
		`## :rotating_light::rotating_light: Bestellwunschaufnahme!!! :rotating_light::rotating_light:
## Gewinner ist %s %s !!!!
<@&%s>

Weiteres wird per slash-command geklärt ("/order") :saluting_face:!

# Alle Bestellungen bitte bis %s Uhr abgeben.
Danach wird **nichts** mehr angenommen.`,
		dm.Name,
		dm.Emoji,
		args.DonerRoles[1],
		expiry.Format("15:04"),
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

func sendPollMessage(s *discordgo.Session, dms []*doner.DonerMan) *discordgo.Message {
	endTime := time.Now().Add(*args.PollDuration).Format("15:04")
	answers := buildVoteAnswers(dms)
	pollMsg, err := s.ChannelMessageSendComplex(*args.DonerChannel, &discordgo.MessageSend{
		Content: fmt.Sprintf(`# 🚨🚨 Welcher Dönermann wird heute beansprucht. 🚨🚨
## Jetzt wird freiheitlich **DEMOKRATISCH** gewählt!!! (Ende %s Uhr)
<@&%s> <@&%s>`, endTime, args.DonerRoles[0], args.DonerRoles[1]),
		AllowedMentions: &discordgo.MessageAllowedMentions{Roles: args.DonerRoles},
		Poll: &discordgo.Poll{
			Question: discordgo.PollMedia{
				Text: fmt.Sprintf("Wähle deinen %s-Fabrikanten des Vertrauens!!!",
					doner.GetRandomDonerName(),
				),
			},
			Answers:    answers,
			LayoutType: discordgo.PollLayoutTypeDefault,
			Duration:   int(args.PollDuration.Hours()) + 1,
		},
	})
	if err != nil {
		log.Fatalln("Could not send poll: " + err.Error())
	}
	log.Println("Poll end: " + endTime)
	return pollMsg
}
