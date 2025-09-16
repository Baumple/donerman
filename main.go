package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID  = flag.String("guild", "", "Test guild ID")
	BotToken = flag.String("token", "", "Bot access token")
	AppID    = flag.String("app", "", "Application ID")
)

func init() {
	flag.Parse()
}

var s discordgo.Session

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"hello": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hello",
					Flags:   discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Yes",
									Style:    discordgo.DangerButton,
									CustomID: "fd_deez",
								},
							},
						},
					},
				},
			})
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	componentHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
)

var (
	DonerRoles = []string{"1371736220748611636", "1417481551830188145"}
)

func startDonerMenPoll(s *discordgo.Session, donerMen []DonerMan) (*discordgo.Message, *time.Timer) {
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
	pollMsg, err := s.ChannelMessageSendComplex("1417472533762015294", &discordgo.MessageSend{
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
		os.Exit(1)
	}
	return pollMsg, time.NewTimer(15 * time.Minute)
}

func startOrder(s *discordgo.Session, pollMsg *discordgo.Message) {
	pollMsg, err := s.ChannelMessage(pollMsg.ChannelID, pollMsg.ID)
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
	}
	answers := pollMsg.Poll.Answers
	results := pollMsg.Poll.Results
	maxVotes := discordgo.PollAnswerCount{
		ID:      0,
		Count:   -1,
		MeVoted: false,
	}
	for _, a := range results.AnswerCounts {
		if a.Count > maxVotes.Count {
			maxVotes = *a
		}
	}
}

func main() {
	dm, err := GetDonerMen()
	if err != nil {
		log.Fatalln("Could not read donermen: " + err.Error())
	}

	s, err := discordgo.New("Bot " + *BotToken)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up")
	})

	// s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 	switch i.Type {
	// 	case discordgo.InteractionApplicationCommand:
	// 		log.Println("hey")
	// 		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
	// 			h(s, i)
	// 		}
	// 	}
	// })

	pollMsg, pollTimer := startDonerMenPoll(s, dm)

	if err = s.Open(); err != nil {
		log.Fatalln("Error: " + err.Error())
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-pollTimer.C:
		log.Println("Poll finished")
		startOrder(s, pollMsg)
	case <-stop:
		log.Println("Shutdown")
	}
}
