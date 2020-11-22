package main

import (
	"log"
	"os"
	"reflect"

	"github.com/diamondburned/arikawa/bot"
)

func main() {
	var token = os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $BOT_TOKEN given.")
	}

	commands := &Bot{}

	wait, err := bot.Start(token, commands, func(ctx *bot.Context) error {
		ctx.HasPrefix = bot.NewPrefix("!")
		ctx.EditableCommands = true
		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	if err := wait(); err != nil {
		log.Fatalln("Gateway fatal error:", err)
	}
}

// itemExists takes a slice of items, and an item, and returns a bool on whether
// that item is in that slice, plus if it is, the index.
func itemExists(slice interface{}, item interface{}) (bool, int) {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true, i
		}
	}

	return false, -1
}

func doExist(s string) bool {
	_, err := os.Stat(s)
	if err != nil {
		return false
	}
	return true
}
