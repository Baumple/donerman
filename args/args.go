package args

import (
	"flag"
	"time"
)

// Bot parameters
func init() {
	flag.Parse()
}

var (
	GuildID      = flag.String("guild", "", "Test guild ID")
	BotToken     = flag.String("token", "", "Bot access token")
	AppID        = flag.String("app", "", "Application ID")
	DonerChannel = flag.String("chan", "", "The channel ID of the order process")

	PollDuration  = flag.Duration("pd", 15*time.Second, "Poll duration")
	OrderDuration = flag.Duration("od", 15*time.Second, "Order duration")
)

var (
	DonerRoles = []string{"1371736220748611636", "1417481551830188145"}
)

