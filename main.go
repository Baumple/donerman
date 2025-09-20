package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
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
	PollDuration = flag.Duration("pd", 15*time.Second, "Poll duration")
)

var s discordgo.Session

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
		log.Println("Poll finished.")
		startOrder(s, pollMsg, dm)
	case <-stop:
		log.Println("Shutdown")
	}
}
