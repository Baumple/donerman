package main

import (
	"flag"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
func init() {
	flag.Parse()
}

var (
	GuildID      = flag.String("guild", "", "Test guild ID")
	BotToken     = flag.String("token", "", "Bot access token")
	AppID        = flag.String("app", "", "Application ID")
	DonerChannel = flag.String("chan", "", "The channel ID of the order process")

	PollDuration  = flag.Duration("pd", 15*time.Second, "Poll duration")
	OrderDuration = flag.Duration("od", 15*time.Second, "Order duration")
)

// TODO: order summary in dm
// TODO: pay in cache and/or paypal

func main() {
	dms, err := GetDonerMen()
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

	if err = s.Open(); err != nil {
		log.Fatalln("Error: " + err.Error())
	}
	defer s.Close()

	dm, voters := startDonerMenPoll(s, dms)
	startOrder(s, dm, voters)

}
