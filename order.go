package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/currency"
)

type Order struct {
	ItemName string
	Price    currency.Amount
	PlacedBy *discordgo.User
}

func getUserOrder(c chan Order) {
}

type orderStage int

const (
	itemName = iota
	itemPrice
	complete
)

type orderState struct {
	user      *discordgo.User
	channel   *discordgo.Channel
	stage     orderStage
	itemName  *string
	itemPrice *float64
}

func createOrderChannels(channelID string, pollID string, answerCounts []*discordgo.PollAnswerCount) []orderState {
	orderStates := []orderState{}
	for _, result := range answerCounts {
		answerVoters, err := s.PollAnswerVoters(channelID, pollID, result.ID)
		if err != nil {
			log.Fatalln("Could not retrieve voters from Poll.")
		}
		for _, voter := range answerVoters {
			if c, err := s.UserChannelCreate(voter.ID); err != nil {
				log.Println("Could not open user channel: ", err.Error())
			} else {
				orderStates = append(
					orderStates,
					orderState{
						user:      voter,
						channel:   c,
						stage:     itemName,
						itemName:  nil,
						itemPrice: nil,
					},
				)
			}
		}
	}
	return orderStates
}

func handleOrderMessage(
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	state *orderState,
	orderChan chan Order,
) {
	switch state.stage {
	case itemName:
		state.itemName = &m.Content
		s.ChannelMessageSend(state.channel.ID, "And how much does it cost?")
		state.stage = itemPrice
	case itemPrice:
		if price, err := strconv.ParseFloat(m.Content, 32); err != nil {
			s.ChannelMessageSend(
				state.channel.ID,
				"Bruh. Ihr und euer scheiß Deutschgeschiss mit dene Dezimalpunkte immer. NUTZT EIN KOMMA WIE ALLE ANDEREN AUCH!1!11!1elf",
			)
			return
		} else {
			state.itemPrice = &price
		}
		orderChan <- Order{
			ItemName: *state.itemName,
			Price:    currency.EUR.Amount(state.itemPrice),
			PlacedBy: state.user,
		}
	}
}

func startOrder(s *discordgo.Session, pollMsg *discordgo.Message, donerMen []DonerMan) {
	log.Println("Starting order.")
	pollMsg, err := s.ChannelMessage(pollMsg.ChannelID, pollMsg.ID)
	if err != nil {
		log.Fatalln(err.Error())
	}

	winner := getWinningDonerMan(donerMen, pollMsg.Poll.Answers, *pollMsg.Poll.Results)
	if winner == nil {
		log.Fatalln("No one voted.")
	}

	var _ = fmt.Sprintf(
		`# :rotating_light::rotating_light: Bestellwunschaufnahme!!! :rotating_light::rotating_light:
## Gewinner ist %s !!!!
<@&%s> <@&%s>

Weiteres wird per DM geklärt :saluting_face:!`,
		winner,
		DonerRoles[0],
		DonerRoles[1],
	)

	orderChan := make(chan Order, 16)
	orderStates := createOrderChannels(pollMsg.ChannelID, pollMsg.ID, pollMsg.Poll.Results.AnswerCounts)
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		for _, state := range orderStates {
			if state.channel.ID == m.ChannelID {
				handleOrderMessage(s, m, &state, orderChan)
				return
			}
		}
	})

	if err != nil {
		log.Fatalln("Error on sending order message: ", err.Error())
	}

	log.Println("Finished order.")
}
