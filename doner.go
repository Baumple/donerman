package main

import "math/rand/v2"

var (
	donerNames = []string{
		"Osmanisches Dreieck",
		"Osmanischer Gulaschkanister",
		"Kalifat-Kalorienrakete",
		"Osmanische Rohrbombe",
		"Türkischer Drehspieß-Propeller",
		"Kreuzritterabwehr-Gulaschwalze",
		"Kreuzritterabwehr-Nährstoffbombe",
		"Baba's Bauchrakete",
		"Sesamsoßen-Sprengsatz",
		"Anatolischer Protein-Torpedo",
		"Anatolische Drehfleisch-Rakete",
		"Fleischfladen-Frachtgeschoss",
		"Knoblauchsaußen-Kavallerie",
		"Lamacun-Lebergranate",
		"Des Kreuzritter's Albtraum",
		"Heiliger Gral der Hackfleischhydraulik",
		"Fleischkelch des Bosporus",
		"Knoblauchtes Darmverschluss Edelgeschoss",
	}
)

func GetRandomDonerName() string {
	return donerNames[rand.IntN(len(donerNames))]
}
