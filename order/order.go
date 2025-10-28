package order

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/baumple/donerman/config"
	"github.com/baumple/donerman/doner"
	"github.com/baumple/donerman/utils"

	"github.com/bwmarrin/discordgo"
)

type Order struct {
	ItemName      string
	PricePerPiece float64
	Amount        int
	Comment       string
	PlacedBy      *discordgo.User
	PaymentMethod PaymentMethod
}

func (o *Order) TotalPrice() float64 {
	return o.PricePerPiece * float64(o.Amount)
}

type PaymentMethod int

const (
	paymentPaypal = PaymentMethod(iota)
	paymentCash
	paymentDebt
)

func (p PaymentMethod) String() string {
	switch p {
	case paymentCash:
		return "Cash"
	case paymentPaypal:
		return "Paypal"
	case paymentDebt:
		return "Debt"
	}
	log.Fatalf("Received illegal payment method: %d", p)
	panic("")
}

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
					Name:        "payment-method",
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
						{
							Name:  "debt",
							Value: paymentDebt,
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
	expiry := time.Now().Local().Add(*config.OrderDuration)
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
	state := orderState{
		orders: make(map[string][]Order),
		users:  make(map[string]*discordgo.User),
		lock:   &sync.RWMutex{},
	}

	timer := time.NewTimer(*config.OrderDuration + 5*time.Second)
	deleteCommands := createCommands(s, timer, &state)
	defer deleteCommands()

	<-timer.C
	log.Println("Finished order.")
	return state.orders, state.users

}

// Creates slash commands and handler.
// Returns a function which deletes the handler and commands it created.
func createCommands(
	s *discordgo.Session,
	orderTimer *time.Timer,
	state *orderState,
) func() {
	removeHandler := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		optionMap := make(
			map[string]*discordgo.ApplicationCommandInteractionDataOption,
			len(options),
		)
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		switch i.ApplicationCommandData().Name {
		case "advance-order":
			orderTimer.Reset(0)
			if utils.IsOberdirektor(i.Member) {
				utils.SendConfirmation(s, i)
			} else {
				utils.SendNotAllowed(s, i)
			}
		case "set-order-time":
			if !utils.IsOberdirektor(i.Member) {
				utils.SendNotAllowed(s, i)
				return
			}
			var minutes time.Duration
			if m, ok := optionMap["minutes"]; ok {
				minutes = time.Duration(m.IntValue())
			}
			var seconds time.Duration
			if s, ok := optionMap["seconds"]; ok {
				seconds = time.Duration(s.IntValue())
			}

			duration := time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
			timeEnd := time.Now().Add(duration)

			orderTimer.Reset(duration)
			log.Println("Order now ends: " + duration.String())

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Bestellung endet nun um: " + timeEnd.Format("15:04"),
				},
			})
			if err != nil {
				log.Println("Could not send confirmation message" + err.Error())
			}

		case "order":
			handleOrderMessage(s, i, optionMap, state)
		}
	})

	cmds := []*discordgo.ApplicationCommand{}
	for _, command := range commands {
		cmd, err := s.ApplicationCommandCreate(*config.AppID, *config.GuildID, command)
		if err != nil {
			log.Fatalln("Could not create order commands: " + err.Error())
		}
		cmds = append(cmds, cmd)
	}

	return func() {
		for _, cmd := range cmds {
			err := s.ApplicationCommandDelete(*config.AppID, *config.GuildID, cmd.ID)
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
	state *orderState,
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
	payMethod, ok := optionMap["payment-method"]
	if !ok {
		log.Println("Not all options were provided")
		return
	}

	amountOpt, ok := optionMap["amount"]
	amount := 1
	if ok {
		amount = int(amountOpt.IntValue())
	}

	state.addUserAndOrder(i.Member.User, Order{
		ItemName:      itemName.StringValue(),
		PricePerPiece: itemPrice.FloatValue(),
		Amount:        amount,
		PlacedBy:      i.Member.User,
		PaymentMethod: PaymentMethod(payMethod.IntValue()),
	})

	log.Printf("Received order: %s by %s (%.02f)",
		itemName.StringValue(), i.Member.User.Username, itemPrice.FloatValue(),
	)

	orders := state.getUserOrders(i.Member.User.ID)

	fields := []*discordgo.MessageEmbedField{}
	sum := 0.0
	for i := range orders {
		order := orders[i]
		f := discordgo.MessageEmbedField{
			Name: fmt.Sprintf(
				"%d x %s€",
				order.Amount,
				order.ItemName,
			),
			Value: fmt.Sprintf("%d x %.02f€ = %.02f€",
				order.Amount,
				order.PricePerPiece,
				order.TotalPrice(),
			),
			Inline: true,
		}
		fields = append(fields, &f)
		sum += order.TotalPrice()
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Insgesamt:",
		Value:  fmt.Sprintf("%.02f€", sum),
		Inline: false,
	})
	embed := discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  "Deine derzeitige Bestellungen",
		Fields: fields,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Bestellung erfolgreich!",
			Embeds:  []*discordgo.MessageEmbed{&embed},
		},
	})
}
