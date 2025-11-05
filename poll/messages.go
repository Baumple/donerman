package poll

import (
	"fmt"
	"log"
	"time"

	"github.com/baumple/donerman/config"
	"github.com/baumple/donerman/doner"
	"github.com/bwmarrin/discordgo"
)

func sendInvalidVoteMessage(s *discordgo.Session) {
	_, err := s.ChannelMessageSendComplex(
		*config.DonerChannel,
		&discordgo.MessageSend{
			Content: fmt.Sprintf("Niemand hat gevoted. Heute gibt es wohl kein %s @here",
				doner.GetRandomDonerName(),
			),
		},
	)
	if err != nil {
		log.Fatalln("Could not send invalid vote message: " + err.Error())
	}
}

func announcePollWinner(s *discordgo.Session, dm *doner.DonerMan, expiry time.Time) {
	var content = fmt.Sprintf(
		`## :rotating_light::rotating_light: Bestellwunschaufnahme!!! :rotating_light::rotating_light:
## Gewinner ist %s %s !!!!
@here

Weiteres wird per slash-command geklärt ("/order") :saluting_face:!

# Alle Bestellungen bitte bis %s Uhr abgeben.
Danach wird **nichts** mehr angenommen.`,
		dm.Name,
		dm.Emoji,
		expiry.Format("15:04"),
	)
	_, err := s.ChannelMessageSendComplex(*config.DonerChannel, &discordgo.MessageSend{
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
	endTime := time.Now().Add(*config.PollDuration).Format("15:04")
	answers := buildVoteAnswers(dms)
	pollMsg, err := s.ChannelMessageSendComplex(*config.DonerChannel, &discordgo.MessageSend{
		Content: fmt.Sprintf(`# 🚨🚨 Welcher Dönermann wird heute beansprucht. 🚨🚨
## Jetzt wird freiheitlich **DEMOKRATISCH** gewählt!!! (Ende %s Uhr)
@here`, endTime),
		Poll: &discordgo.Poll{
			Question: discordgo.PollMedia{
				Text: fmt.Sprintf("Wähle deinen %s-Fabrikanten des Vertrauens!!!",
					doner.GetRandomDonerName(),
				),
			},
			Answers:    answers,
			LayoutType: discordgo.PollLayoutTypeDefault,
			Duration:   int(config.PollDuration.Hours()) + 1,
		},
	})
	if err != nil {
		log.Fatalln("Could not send poll: " + err.Error())
	}
	log.Println("Poll end: " + endTime)
	return pollMsg
}
