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

func StartDonerMenPoll(s *discordgo.Session, dms []*doner.DonerMan) (*doner.DonerMan, []*discordgo.User) {
	log.Println("Starting poll.")

	answers := buildVoteAnswers(dms)
	pollMsg, err := s.ChannelMessageSendComplex(*args.DonerChannel, &discordgo.MessageSend{
		Content: fmt.Sprintf(`# 🚨🚨 Welcher Dönermann wird heute beansprucht. 🚨🚨
## Jetzt wird freiheitlich **DEMOKRATISCH** gewählt!!!
<@&%s> <@&%s>`, args.DonerRoles[0], args.DonerRoles[1]),
		AllowedMentions: &discordgo.MessageAllowedMentions{Roles: args.DonerRoles},
		Poll: &discordgo.Poll{
			Question: discordgo.PollMedia{
				Text: fmt.Sprintf("Wähle deinen %s-Fabrikanten des Vertrauens!!!",
					doner.GetRandomDonerName(),
				),
			},
			Answers:          answers,
			AllowMultiselect: false,
			LayoutType:       1,
			Duration:         1,
		},
	})
	if err != nil {
		log.Fatalln("Could not send poll: " + err.Error())
	}

	pollTimer := time.NewTimer(*args.PollDuration)
	<-pollTimer.C

	log.Println("Poll finished.")
	updatePoll(s, pollMsg)

	winner := getPollWinner(s, pollMsg, dms)
	users := getPollVoters(s, pollMsg)

	announcePollWinner(s, winner)
	return winner, users
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

func getPollVoters(s *discordgo.Session, pollMsg *discordgo.Message) []*discordgo.User {
	users := []*discordgo.User{}
	for _, result := range pollMsg.Poll.Results.AnswerCounts {
		answerVoters, err := s.PollAnswerVoters(pollMsg.ChannelID, pollMsg.ID, result.ID)
		if err != nil {
			log.Fatalln("Could not retrieve voters from Poll.")
		} else {
			for _, voter := range answerVoters {
				users = append(users, voter)
			}
		}
	}
	return users
}

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

func getPollWinner(s *discordgo.Session, pollMsg *discordgo.Message, donerMen []*doner.DonerMan) *doner.DonerMan {
	answers := pollMsg.Poll.Answers
	results := pollMsg.Poll.Results
	if len(donerMen) <= 0 {
		log.Panicln("No doner men")
	}
	if len(results.AnswerCounts) <= 0 {
		sendInvalidVoteMessage(s)
	}

	maxVotes := results.AnswerCounts[0]

	// get max voted answer id
	for _, a := range results.AnswerCounts {
		if a.Count > maxVotes.Count {
			maxVotes = a
		}
	}

	// match answer id to actual answer, and then answer to doner man
	for _, answer := range answers {
		if answer.AnswerID == maxVotes.ID {
			for _, dm := range donerMen {
				if dm.Name == answer.Media.Text {
					return dm
				}
			}
		}
	}
	panic("This should be unreachable")
}

func buildVoteAnswers(dms []*doner.DonerMan) []discordgo.PollAnswer {
	answers := []discordgo.PollAnswer{}
	for _, donerMan := range dms {
		media := discordgo.PollMedia{
			Text:  donerMan.Name,
			Emoji: &discordgo.ComponentEmoji{Name: "📜"},
		}
		option := discordgo.PollAnswer{
			Media: &media,
		}
		answers = append(answers, option)
	}
	return answers
}

// fetches the latest state of the given msg (including poll votes)
func updatePoll(s *discordgo.Session, pollMsg *discordgo.Message) {
	updated, err := s.ChannelMessage(pollMsg.ChannelID, pollMsg.ID)
	if err != nil {
		log.Fatalln("Could not get vote results: " + err.Error())
	}
	*pollMsg = *updated
}
