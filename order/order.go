package order

import (
	"log"
	"strconv"
	"time"

	"github.com/baumple/donerman/args"
	"github.com/baumple/donerman/doner"
	"github.com/bwmarrin/discordgo"
)

type Order struct {
	ItemName string
	Price    float64
	PlacedBy *discordgo.User
}

type orderStage int

const (
	itemNameStage = orderStage(iota)
	itemPriceStage
	completeStage
)

type orderState struct {
	user        *discordgo.User
	userChannel *discordgo.Channel
	stage       orderStage
	itemName    *string
	itemPrice   *float64
}

func (o *orderState) ResetOrder() {
	o.itemName = new(string)
	o.itemPrice = new(float64)
	o.stage = itemNameStage
}

func (o *orderState) isValid() bool {
	return o.itemName != nil && o.itemPrice != nil
}

func StartOrder(s *discordgo.Session, dm *doner.DonerMan, voters []*discordgo.User) map[*discordgo.User][]Order {
	log.Println("Starting order.")
	orders := getOrdersFromUsers(s, dm, voters)
	sendOrderSummary(s, orders)
	return orders
}

// initOrderChannels initiates DM channels to each discordgo.User, who has voted in the poll
//
// `channelID` is the ID channel of the poll,
// `pollID` is the ID of the poll, while answerCounts are contains the vote results.
//
// Returns a map of userId -> `orderStates`.
func initOrderChannels(
	s *discordgo.Session,
	voters []*discordgo.User,
) map[string]*orderState {
	log.Println("Creating user channels.")
	userOrderStateMap := make(map[string]*orderState)

	for _, voter := range voters {
		userChannel, err := s.UserChannelCreate(voter.ID)
		if err != nil {
			log.Println("Could not open channel to user: " + err.Error())
		}

		userOrderStateMap[voter.ID] = &orderState{
			user:        voter,
			userChannel: userChannel,
			stage:       itemNameStage,
			itemName:    nil,
			itemPrice:   nil,
		}
	}
	return userOrderStateMap
}

// WARN:
// TODO: Find another way instead of pointer to orders map
func getOrdersFromUsers(
	s *discordgo.Session,
	dm *doner.DonerMan,
	voters []*discordgo.User,
) map[*discordgo.User][]Order {
	userOrderStateMap := initOrderChannels(s, voters)

	orderChan := make(chan (Order), 16)
	removeHandler := s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if state, ok := userOrderStateMap[m.Author.ID]; ok {
			handleOrderMessage(s, m, state, orderChan)
		}
	})

	expiry := time.Now().Add(*args.OrderDuration)
	timer := time.NewTimer(*args.OrderDuration)

	for _, o := range userOrderStateMap {
		err := sendOrderMessage(s, dm, o, expiry)
		if err != nil {
			log.Println("Error on sending order message: ", err.Error())
		}
	}

	orders := make(map[*discordgo.User][]Order)
	for {
		select {
		case order := <-orderChan:
			orders[order.PlacedBy] = append(orders[order.PlacedBy], order)

		case <-timer.C:
			log.Println("Finished order.")
			removeHandler()
			return orders
		}
	}

}

func handleOrderMessage(
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	state *orderState,
	orderChan chan Order,
) {
	switch state.stage {
	case itemNameStage:
		state.itemName = new(string)
		*state.itemName = m.Content

		_, err := s.ChannelMessageSend(
			state.userChannel.ID,
			"Und wie viel kostet das?\n(Bitte nur Zahlen mit **DEZIMALPUNKT** und ohne sonstigen Firlefanz, "+
				"sonst dreh ich hier durch und du bekommst nichts als nen feuchten Furz bar von meiner Hand)",
		)
		if err != nil {
			log.Printf("Unable to send followup message asking for item price to user '%s': %s\n",
				state.user.Username, err.Error(),
			)
		}

		state.stage = itemPriceStage
		log.Printf("Received item name from %s: %s", state.user.Username, *state.itemName)

	case itemPriceStage:
		if price, err := strconv.ParseFloat(m.Content, 32); err != nil {
			s.ChannelMessageSend(
				state.userChannel.ID,
				"Bruh. WAS HAB ICH DENN GERADE EBEN GESAGT!!111",
			)
			return
		} else {
			state.itemPrice = &price
			state.stage = completeStage
		}

		if state.isValid() {
			log.Printf("Received item price from %s: %f", state.user.Username, *state.itemPrice)
			_, err := s.ChannelMessageSend(state.userChannel.ID,
				"Bestellung wurde aufgenommen.\n"+
					"Wenn du noch etwas bestellen willst, schreib einfach nochmal.")
			if err != nil {
				log.Println("Could not send follow up message" + err.Error())
			}

			orderChan <- Order{
				ItemName: *state.itemName,
				Price:    *state.itemPrice,
				PlacedBy: state.user,
			}

			state.ResetOrder()

		} else {
			log.Printf("Got invalid order: %#v\n", state)
		}
	}
}
