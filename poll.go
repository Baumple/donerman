package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	DonerRoles = []string{"1371736220748611636", "1417481551830188145"}
)

func startDonerMenPoll(s *discordgo.Session, donerMen []DonerMan) (*discordgo.Message, *time.Timer) {
	log.Println("Starting poll.")

	options := []discordgo.PollAnswer{}
	for _, donerMan := range donerMen {
		media := discordgo.PollMedia{
			Text:  donerMan.Name,
			Emoji: &discordgo.ComponentEmoji{Name: "📜"},
		}
		option := discordgo.PollAnswer{
			Media: &media,
		}
		options = append(options, option)
	}
	pollMsg, err := s.ChannelMessageSendComplex(*DonerChannel, &discordgo.MessageSend{
		Content: fmt.Sprintf(`# 🚨🚨 Welcher Dönermann wird heute beansprucht. 🚨🚨
## Jetzt wird freiheitlich **DEMOKRATISCH** gewählt!!!
<@&%s> <@&%s>`, DonerRoles[0], DonerRoles[1]),
		AllowedMentions: &discordgo.MessageAllowedMentions{Roles: DonerRoles},
		Poll: &discordgo.Poll{
			Question: discordgo.PollMedia{
				Text: "Wähle deinen Osmanischen Gulaschkanister Fabrikanten des Vertrauens!!!",
			},
			Answers:          options,
			AllowMultiselect: false,
			LayoutType:       1,
			Duration:         1,
		},
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	return pollMsg, time.NewTimer(*PollDuration)
}

func getWinningDonerMan(
	donerMen []DonerMan,
	answers []discordgo.PollAnswer,
	results discordgo.PollResults,
) *DonerMan {
	if len(donerMen) <= 0 {
		return nil
	}
	maxVotes := results.AnswerCounts[0]
	for _, a := range results.AnswerCounts {
		if a.Count > maxVotes.Count {
			maxVotes = a
		}
	}
	for _, answer := range answers {
		if answer.AnswerID == maxVotes.ID {
			for _, dm := range donerMen {
				if dm.Name == answer.Media.Text {
					return &dm
				}
			}
		}
	}
	panic("This should be unreachable")
}

