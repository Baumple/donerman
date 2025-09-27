package poll

import (
	"log"
	"time"

	"github.com/baumple/donerman/args"
	"github.com/baumple/donerman/doner"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "advance-poll",
			Description: "Zur Bestellungsphase übergehen",
		},
		{
			Name:        "set-poll-time",
			Description: "Umfragedauer neu setzen",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "Minuten",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "Sekunden",
					Required:    false,
				},
			},
		},
	}
)

// Returns a function which deletes both the handler and the commands
// TODO: confirmation messages
func createCommands(s *discordgo.Session, pollTimer *time.Timer) func() {
	removeHandler := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optMap := make(
			map[string]*discordgo.ApplicationCommandInteractionDataOption,
			len(i.ApplicationCommandData().Options),
		)
		for _, opt := range i.ApplicationCommandData().Options {
			optMap[opt.Name] = opt
		}

		switch i.ApplicationCommandData().Name {
		case "advance-poll":
			pollTimer.Reset(0)
		case "set-poll-time":
			var minutes time.Duration
			if m, ok := optMap["minutes"]; ok {
				minutes = time.Duration(m.IntValue())
			}
			var seconds time.Duration
			if s, ok := optMap["seconds"]; ok {
				seconds = time.Duration(s.IntValue())
			}

			pollTimer.Reset(time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second)
		}
	})

	log.Println("Adding commands")

	cmds := []*discordgo.ApplicationCommand{}
	for _, cmd := range commands {
		cmd, err := s.ApplicationCommandCreate(*args.AppID, *args.GuildID, cmd)
		if err != nil {
			log.Fatalln("Could not create poll commands:" + err.Error())
		}
		cmds = append(cmds, cmd)
	}

	return func() {
		log.Println("Deleting commands")
		for _, cmd := range cmds {
			err := s.ApplicationCommandDelete(*args.AppID, *args.GuildID, cmd.ID)
			if err != nil {
				log.Println("Could not delete poll commands: " + err.Error())
			}
		}
		removeHandler()
	}
}

func StartDonerMenPoll(s *discordgo.Session, dms []*doner.DonerMan) (*doner.DonerMan, []*discordgo.User) {
	log.Println("Starting poll.")

	pollMsg := sendPollMessage(s, dms)

	pollTimer := time.NewTimer(*args.PollDuration)

	deleteCommands := createCommands(s, pollTimer)
	defer deleteCommands()

	<-pollTimer.C

	log.Println("Poll finished.")

	updatePoll(s, pollMsg)

	winner := getPollWinner(s, pollMsg, dms)
	users := getPollVoters(s, pollMsg)

	expiry := time.Now().Add(*args.OrderDuration).Local()
	announcePollWinner(s, winner, expiry)
	return winner, users
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
