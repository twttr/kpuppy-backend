package domain

import (
	"encoding/hex"
	"fmt"
)

var adjectives = []string{
	"Pink", "Blue", "Green", "Purple", "Orange", "Red", "Yellow", "Silver",
	"Golden", "Brave", "Swift", "Calm", "Wise", "Happy", "Lucky", "Gentle",
	"Bold", "Clever", "Quiet", "Bright", "Noble", "Merry", "Proud", "Shy",
	"Keen", "Wild", "Lazy", "Tiny", "Giant", "Ancient", "Young", "Cosmic",
}

var animals = []string{
	"Koala", "Tiger", "Panda", "Eagle", "Wolf", "Fox", "Bear", "Lion",
	"Owl", "Hawk", "Deer", "Rabbit", "Dolphin", "Whale", "Penguin", "Seal",
	"Otter", "Beaver", "Badger", "Raccoon", "Squirrel", "Hedgehog", "Parrot", "Falcon",
	"Lynx", "Jaguar", "Panther", "Leopard", "Cheetah", "Gazelle", "Moose", "Bison",
}

func GeneratePseudonym(userHash string) string {
	bytes, err := hex.DecodeString(userHash)
	if err != nil || len(bytes) < 3 {
		return "Anonymous User"
	}
	adjIndex := int(bytes[0]) % len(adjectives)
	animalIndex := int(bytes[1]) % len(animals)
	suffix := int(bytes[2]) % 100
	return fmt.Sprintf("%s %s %d", adjectives[adjIndex], animals[animalIndex], suffix)
}
