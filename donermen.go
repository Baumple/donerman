package main

import (
	"encoding/json"
	"os"
)

type DonerMan struct {
	Name    string `json:"name"`
	MenuURL string `json:"menu_url"`
	Emoji   string `json:"emoji"`
}

type donerMen struct {
	Donermen []DonerMan
}

func GetDonerMen() ([]DonerMan, error) {
	f, err := os.ReadFile("./donermen.json")
	if err != nil {
		return nil, err
	}
	dm := donerMen{}
	err = json.Unmarshal(f, &dm)
	if err != nil {
		return nil, err
	}
	return dm.Donermen, nil
}
