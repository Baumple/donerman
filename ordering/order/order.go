package order

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type Order struct {
	ItemName      string
	PricePerPiece float64
	Amount        int
	Comment       string
	PlacedBy      *discordgo.User
	PaymentMethod PaymentMethod
	ID            uuid.UUID
}

func (o *Order) TotalPrice() float64 {
	return o.PricePerPiece * float64(o.Amount)
}

type PaymentMethod int

const (
	PaymentPaypal = PaymentMethod(iota)
	PaymentCash
	PaymentDebt
)

func (p PaymentMethod) String() string {
	switch p {
	case PaymentCash:
		return "Cash"
	case PaymentPaypal:
		return "Paypal"
	case PaymentDebt:
		return "Debt"
	}
	log.Fatalf("Received illegal payment method: %d", p)
	panic("")
}
