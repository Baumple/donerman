package main

import (
	"log"

	"github.com/baumple/donerman/args"
	"github.com/baumple/donerman/doner"
	"github.com/baumple/donerman/order"
	"github.com/baumple/donerman/poll"

	"github.com/bwmarrin/discordgo"
)

// TODO: order summary in dm
// TODO: Special offers
// TODO: pickup or delivery
// TODO: more detailed output
//       - log expected poll/order time end
// TODO: config for announcements

func main() {
	dms, err := doner.GetDonerMen()
	if err != nil {
		log.Fatalln("Could not read donermen: " + err.Error())
	}

	s, err := discordgo.New("Bot " + *args.BotToken)
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
