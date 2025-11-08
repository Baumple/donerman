package orderstate

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/baumple/donerman/config"
	"github.com/baumple/donerman/ordering/order"
)

type OrderState struct {
	orders     map[string][]order.Order
	users      map[string]*discordgo.User
	OrderTimer *time.Timer
	lock       *sync.RWMutex
}

func New() *OrderState {
	return &OrderState{
		orders:     make(map[string][]order.Order),
		users:      make(map[string]*discordgo.User),
		lock:       &sync.RWMutex{},
		OrderTimer: time.NewTimer(*config.OrderDuration + 5*time.Second),
	}
}

func (o *OrderState) ResetOrders(userID string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.orders[userID] = []order.Order{}
}

func (o *OrderState) RemoveOrder(userID string, orderID string) bool {
	o.lock.Lock()
	defer o.lock.Unlock()

	currentOrders := o.orders[userID]
	newOrders := []order.Order{}

	found := false
	for i := range currentOrders {
		if currentOrders[i].ID != uuid.MustParse(orderID) {
			newOrders = append(newOrders, currentOrders[i])
		} else {
			found = true
		}
	}

	o.orders[userID] = newOrders
	return found
}

func (o *OrderState) GetUserOrders(id string) []order.Order {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.orders[id]
}

func (o *OrderState) AddUserAndOrder(user *discordgo.User, order order.Order) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.users[user.ID]; !ok {
		o.users[user.ID] = user
	}
	o.orders[user.ID] = append(o.orders[user.ID], order)
}

func (o *OrderState) AwaitFinish() (map[string][]order.Order, map[string]*discordgo.User) {
	<-o.OrderTimer.C
	return o.orders, o.users
}
