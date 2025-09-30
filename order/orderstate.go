package order

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type orderState struct {
	orders map[string][]Order
	users  map[string]*discordgo.User
	lock   *sync.RWMutex
}

func (o *orderState) getUserOrders(id string) []Order {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.orders[id]
}

func (o *orderState) addUserAndOrder(user *discordgo.User, order Order) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.users[user.ID]; !ok {
		o.users[user.ID] = user
	}
	o.orders[user.ID] = append(o.orders[user.ID], order)
}
