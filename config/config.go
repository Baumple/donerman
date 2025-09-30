package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/baumple/donerman/doner"
)

// TODO: add duration fields
type config struct {
	DonerManRole     string `json:"donerman_role"`
	OberdirektorRole string `json:"oberdirektor_role"`

	DonerMen []doner.DonerMan `json:"donermen"`
}

var C config

var (
	DonerManRole     = ""
	OberdirektorRole = ""
	OberdirektorUser = "378177170083807233"
)

func init() {
	c, err := os.ReadFile("./donermen.json")
	if err != nil {
		log.Fatalln("Could not read donermen.json file: " + err.Error())
	}

	var con config
	if err = json.Unmarshal(c, &con); err != nil {
		log.Fatalln("Error while parsing config: " + err.Error())
	}

	DonerManRole = con.DonerManRole
	OberdirektorRole = con.OberdirektorRole
}

var (
	GuildID       = flag.String("guild", "", "Test guild ID")
	BotToken      = flag.String("token", "", "Bot access token")
	AppID         = flag.String("app", "", "Application ID")
	DonerChannel  = flag.String("chan", "", "The channel ID of the order process")
	PollDuration  = flag.Duration("pd", 2*time.Hour, "Poll duration")
	OrderDuration = flag.Duration("od", 1*time.Hour+50*time.Minute, "Order duration")
	Until         = flag.String(
		"until",
		"",
		"Time when order is ended (poll and order time will be split equally)",
	)
)

// Bot parameters
func init() {
	flag.Parse()

	if *Until != "" {
		t, err := time.Parse("15:04", *Until)

		if err != nil {
			log.Fatalln("Invalid time format passed to `until`: " + err.Error())
		}

		today := time.Now()
		duration := time.Duration(t.Hour()-today.Hour())*time.Hour +
			time.Duration(t.Minute()-today.Minute())*time.Minute

		*PollDuration = duration / 2
		*OrderDuration = duration / 2
	}

}
