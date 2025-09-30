package main

import (
	"log"

	"github.com/baumple/donerman/config"
	"github.com/baumple/donerman/doner"
	"github.com/baumple/donerman/order"
	"github.com/baumple/donerman/poll"

	"github.com/bwmarrin/discordgo"
)

// TODO: remove order command
// TODO: special offers
// TODO: optional comment on order
// TODO: pickup or delivery
// TODO: config for announcements
// TODO: roles per server config

func main() {
	dms, err := doner.GetDonerMen()
	if err != nil {
		log.Fatalln("Could not read donermen: " + err.Error())
	}

	s, err := discordgo.New("Bot " + *config.BotToken)

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up")
	})

	if err = s.Open(); err != nil {
		log.Fatalln("Error: " + err.Error())
	}
	defer s.Close()

	dm, voters := poll.StartDonerMenPoll(s, dms)
	order.StartOrder(s, dm, voters)

}
