package doner

import "math/rand/v2"

var (
	donerNames = []string{
		"Osmanisches Dreieck",
		"Osmanischer Gulaschkanister",
		"Kalifat-Kalorienrakete",
		"Osmanische Rohrbombe",
		"Türkischer Drehspieß-Propeller",
		"Kreuzritterabwehr-Gulaschwalze",
		"Kreuzritterabwehr-Kalifats-Kalorienrakete (KKK)",
		"Baba's Bauchrakete",
		"Sesamsoßen-Sprengsatz",
		"Anatolischer Protein-Torpedo",
		"Anatolische Drehfleisch-Rakete",
		"Fleischfladen-Frachtgeschoss",
		"Knoblauchsaußen-Kavallerie",
		"Lamacun-Lebergranate",
		"Heiliger Gral der Hackfleischhydraulik",
		"Fleischkelch des Bosporus",
		"Knoblauchtes Darmverschluss 50mm-Edelgeschoss",
		"Treibspiegelknoblauchgranate",
		"sechste des Säule des Islams",
	}
)

func GetRandomDonerName() string {
	return donerNames[rand.IntN(len(donerNames))]
}
