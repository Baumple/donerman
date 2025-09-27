package order

import (
	"log"
	"time"

	"github.com/baumple/donerman/args"
	"github.com/baumple/donerman/doner"
	"github.com/bwmarrin/discordgo"
)

type Order struct {
	ItemName      string
	PricePerPiece float64
	Amount        int
	PlacedBy      *discordgo.User
	PaymentMethod PaymentMethod
}

func (o *Order) TotalPrice() float64 {
	return o.PricePerPiece * float64(o.Amount)
}

type PaymentMethod int

func (p PaymentMethod) String() string {
	if p == paymentCash {
		return "Paypal"
	} else if p == paymentPaypal {
		return "Cash"
	}
	log.Fatalf("Received illegal payment method: %d", p)
	panic(nil)
}

const (
	paymentPaypal = PaymentMethod(iota)
	paymentCash
)

var (
	minPrice  = 0.01
	minAmount = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "order",
			Description: "Eine Bestellung aufgeben",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "item-name",
					Description: "Was möchtest du bestellen?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "item-price",
					Description: "Preis pro Stück (Euro)",
					Required:    true,
					MinValue:    &minPrice,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "cash-or-paypal",
					Description: "Bar oder mit Karte (paypal)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "bar",
							Value: paymentCash,
						},
						{
							Name:  "paypal",
							Value: paymentPaypal,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "Wie viele?",
					Required:    false,
					MinValue:    &minAmount,
				},
			},
		},
		{
			Name:        "advance-order",
			Description: "Bestellungen vorzeitig beenden und Bestellzusammenfassung schicken",
		},
		{
			Name:        "set-order-time",
			Description: "Bestelldauer neu setzen",
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

func StartOrder(s *discordgo.Session, dm *doner.DonerMan, voters []*discordgo.User) map[string][]Order {
	log.Println("Starting order.")
	expiry := time.Now().Local().Add(*args.OrderDuration)
	log.Println("Expected order end: " + expiry.Format("15:04"))

	orders, users := getOrdersFromUsers(s, dm, voters)
	sendOrderSummary(s, orders, users)

	return orders
}

func getOrdersFromUsers(
	s *discordgo.Session,
	dm *doner.DonerMan,
	voters []*discordgo.User,
) (map[string][]Order, map[string]*discordgo.User) {
	orderChan := make(chan (Order), 16)

	timer := time.NewTimer(*args.OrderDuration + 5*time.Second)
	deleteCommands := createCommands(s, timer, orderChan)

	defer deleteCommands()

	orders := make(map[string][]Order)
	users := make(map[string]*discordgo.User)
	for {
		select {
		case order := <-orderChan:
			orders[order.PlacedBy.ID] = append(orders[order.PlacedBy.ID], order)
			users[order.PlacedBy.ID] = order.PlacedBy

		case <-timer.C:
			log.Println("Finished order.")
			return orders, users
		}
	}

}

// Creates slash commands and handler.
// Returns a function which deletes the handler and commands it created.
// TODO: confirmation messages
func createCommands(
	s *discordgo.Session,
	orderTimer *time.Timer,
	orderChan chan Order,
) func() {
	removeHandler := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		switch i.ApplicationCommandData().Name {
		case "advance-order":
			orderTimer.Reset(0)
		case "set-order-time":
			var minutes time.Duration
			if m, ok := optionMap["minutes"]; ok {
				minutes = time.Duration(m.IntValue())
			}
			var seconds time.Duration
			if s, ok := optionMap["seconds"]; ok {
				seconds = time.Duration(s.IntValue())
			}

			orderTimer.Reset(time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second)

		case "oder":
			handleOrderMessage(s, i, optionMap, orderChan)
		}
	})

	cmds := []*discordgo.ApplicationCommand{}
	for _, command := range commands {
		cmd, err := s.ApplicationCommandCreate(*args.AppID, *args.GuildID, command)
		if err != nil {
			log.Fatalln("Could not create order commands: " + err.Error())
		}
		cmds = append(cmds, cmd)
	}

	return func() {
		for _, cmd := range cmds {
			err := s.ApplicationCommandDelete(*args.AppID, *args.GuildID, cmd.ID)
			if err != nil {
				log.Println("Could not remove order commands: " + err.Error())
			}
		}
		removeHandler()
	}
}

func handleOrderMessage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
	orderChan chan Order,
) {
	itemName, ok := optionMap["item-name"]
	if !ok {
		log.Println("Not all options were provided")
		return
	}
	itemPrice, ok := optionMap["item-price"]
	if !ok {
		log.Println("Not all options were provided")
		return
	}
	payViaPaypal, ok := optionMap["cash-or-paypal"]
	if !ok {
		log.Println("Not all options were provided")
		return
	}

	amountOpt, ok := optionMap["amount"]
	amount := 1
	if ok {
		amount = int(amountOpt.IntValue())
	}

	orderChan <- Order{
		ItemName:      itemName.StringValue(),
		PricePerPiece: itemPrice.FloatValue(),
		Amount:        amount,
		PlacedBy:      i.Member.User,
		PaymentMethod: PaymentMethod(payViaPaypal.IntValue()),
	}

	log.Printf("Received order: %s by %s (%.02f)",
		itemName.StringValue(), i.Member.User.Username, itemPrice.FloatValue(),
	)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Bestellung erfolgreich!",
			Title:   "Döner",
		},
	})
}
