package ordering

import (
	"fmt"
	"log"
	"time"

	"github.com/baumple/donerman/config"
	"github.com/baumple/donerman/doner"
	"github.com/baumple/donerman/ordering/order"
	"github.com/baumple/donerman/ordering/orderstate"
	"github.com/google/uuid"

	"github.com/bwmarrin/discordgo"
)

var (
	minEuro   = 0.0
	minCents  = 0.0
	maxCents  = 99.0
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
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "item-price-euro",
					Description: "Preis (Euro)",
					Required:    true,
					MinValue:    &minEuro,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "item-price-cent",
					Description: "Preis (cent)",
					Required:    true,
					MaxValue:    maxCents,
					MinValue:    &minCents,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "payment-method",
					Description: "Bar oder mit Karte (paypal)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "bar",
							Value: order.PaymentCash,
						},
						{
							Name:  "paypal",
							Value: order.PaymentPaypal,
						},
						{
							Name:  "debt",
							Value: order.PaymentDebt,
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
			Name:        "edit-orders",
			Description: "Bestellungen bearbeiten",
		},
		{
			Name:        "reset-orders",
			Description: "Bestellungen zurücksetzen",
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

	commandHandlers = map[string]func(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
		state *orderstate.OrderState,
		optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
	){
		"advance-order":  advanceOrder,
		"set-order-time": setOrderTime,

		"reset-orders": func(
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
			state *orderstate.OrderState,
			optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
		) {
			state.ResetOrders(i.Member.User.ID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Deine Bestellungen wurden entfernt."},
			})
		},

		"edit-orders": func(
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
			state *orderstate.OrderState,
			optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
		) {
			handleEditOrder(s, i, state)
		},

		"order": func(
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
			state *orderstate.OrderState,
			optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
		) {
			handleOrder(s, i, optionMap, state)
		},
	}

	removeOrderHandler = func(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
		state *orderstate.OrderState,
	) {
		orderID := i.MessageComponentData().CustomID
		state.RemoveOrder(i.Member.User.ID, orderID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Entfernt!",
			},
		})
	}
)

func StartOrder(s *discordgo.Session, dm *doner.DonerMan, voters []*discordgo.User) map[string][]order.Order {
	log.Println("Starting order.")
	expiry := time.Now().Local().Add(*config.OrderDuration)
	log.Println("Expected order end: " + expiry.Format("15:04"))

	orders, users := getOrdersFromUsers(s)
	sendOrderSummary(s, orders, users)

	return orders
}

func getOrdersFromUsers(
	s *discordgo.Session,
) (map[string][]order.Order, map[string]*discordgo.User) {
	state := orderstate.New()

	deleteCommands := createCommands(s, state)
	defer deleteCommands()

	orders, users := state.AwaitFinish()
	log.Println("Finished order.")
	return orders, users
}

// createCommands creates the slash commands and their handler.
// Returns a function which deletes the handler and commands it created.
func createCommands(
	s *discordgo.Session,
	state *orderstate.OrderState,
) func() {
	removeHandler := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			options := i.ApplicationCommandData().Options
			optionMap := make(
				map[string]*discordgo.ApplicationCommandInteractionDataOption,
				len(options),
			)
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i, state, optionMap)
			}

		case discordgo.InteractionMessageComponent:
			removeOrderHandler(s, i, state)
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

func createOrderContainer(orders []order.Order) []discordgo.MessageComponent {
	containers := []discordgo.MessageComponent{}

	for i := range orders {
		order := orders[i]
		containers = append(containers, &discordgo.Container{
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{Content: fmt.Sprintf(
					`Name: %s
Preis: %d x %.02f = %.02f`,
					order.ItemName,
					order.Amount,
					order.PricePerPiece,
					order.TotalPrice(),
				)},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Entfernen",
							CustomID: order.ID.String(),
						},
					},
				},
			},
		})
	}

	return containers
}

func handleEditOrder(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	state *orderstate.OrderState,
) {
	userOrders := state.GetUserOrders(i.Member.User.ID)
	containers := createOrderContainer(userOrders)

	if len(containers) == 0 {
		containers = []discordgo.MessageComponent{
			discordgo.Container{
				Components: []discordgo.MessageComponent{
					discordgo.TextDisplay{Content: "Keine Bestellungs :/"},
				},
			},
		}
	} else if len(containers) > 8 {
		containers = containers[:8]
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Title: "Deine derzeitigen Bestellungen",
			Flags: discordgo.MessageFlagsIsComponentsV2,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Users: []string{i.Member.User.ID},
			},
			Components: containers,
		},
	})

	if err != nil {
		log.Println("Could not send edit-orders response: " + err.Error())
	}
}

func handleOrder(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
	state *orderstate.OrderState,
) {
	if len(state.GetUserOrders(i.Member.User.ID)) >= 8 {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Du hast schon 8 Bestellungen. Mehr gehen leider nicht.",
			},
		})
		if err != nil {
			log.Println("Could not send 'too many orders' message: " + err.Error())
		}
		return
	}

	itemName, ok := optionMap["item-name"]
	if !ok {
		log.Println("Not all options were provided")
		return
	}
	itemPriceEuro, ok := optionMap["item-price-euro"]
	if !ok {
		log.Println("Not all options were provided")
		return
	}
	itemPriceCents, ok := optionMap["item-price-cent"]
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

	price := float64(itemPriceEuro.IntValue()) +
		(float64(itemPriceCents.IntValue()) / 100.0)
	state.AddUserAndOrder(i.Member.User, order.Order{
		ItemName:      itemName.StringValue(),
		PricePerPiece: price,
		Amount:        amount,
		PlacedBy:      i.Member.User,
		PaymentMethod: order.PaymentMethod(payMethod.IntValue()),
		ID:            uuid.New(),
	})

	log.Printf("Received order: %s by %s (%.02f)",
		itemName.StringValue(), i.Member.User.Username, price,
	)

	orders := state.GetUserOrders(i.Member.User.ID)

	fields := []*discordgo.MessageEmbedField{}
	sum := 0.0
	for i := range orders {
		order := orders[i]
		f := discordgo.MessageEmbedField{
			Name: fmt.Sprintf(
				"%d x %s",
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
