package main

import (
	"log"

	"github.com/diamondburned/arikawa/bot"
	"github.com/diamondburned/arikawa/bot/extras/middlewares"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

// Bot struct
type Bot struct {
	// Context must not be embedded.
	Ctx *bot.Context
	P   *Purger
}

// Setup bot
func (bot *Bot) Setup(sub *bot.Subcommand) {
	me, err := bot.Ctx.Me()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("( Initializing ... ) [ %v is alive ]", me.Username)
	guilds, err := bot.Ctx.Guilds()
	if err != nil {
		log.Fatalln(err)
	}
	for _, g := range guilds {
		log.Printf("( %v ) [ %v bot joined server %v ]", g.Name, me.Username, g.Name)
	}
	// Only allow people in guilds to run guildInfo.
	//sub.AddMiddleware("GuildInfo", middlewares.GuildOnly(bot.Ctx))
	sub.AddMiddleware("Purge", middlewares.AdminOnly(bot.Ctx))
	bot.P = NewPurger()
	err = bot.P.loadChannelList()
	if err != nil {
		log.Println("Couln't load", clf, " - Error:", err)
	}
	if len(bot.P.chids) > 0 {
		for _, c := range bot.P.chids {
			bot.P.last[c] = 0
			log.Printf("( %v ) [ Loaded %v into Purger ]", bot.gInfo(c).Name, bot.cInfo(c).Name)
		}
	}
	go bot.runPurger(bot.P)
}

// Help prints the default help message.
func (bot *Bot) Help(*gateway.MessageCreateEvent) (string, error) {
	return bot.Ctx.Help(), nil
}

// Ping is a simple ping example, perhaps the most simple you could make it.
func (bot *Bot) Ping(*gateway.MessageCreateEvent) (string, error) {
	return "Pong!", nil
}

// gInfo takes a ChannelID and reutns a Guild object from Ctx
func (bot *Bot) gInfo(c discord.ChannelID) *discord.Guild {
	var g *discord.Guild
	g, _ = bot.Ctx.Guild(bot.cInfo(c).GuildID)
	return g
}

// gInfo takes a ChannelID and reutns a Guild object from Ctx
func (bot *Bot) cInfo(c discord.ChannelID) *discord.Channel {
	chInfo, _ := bot.Ctx.Channel(c)
	return chInfo
}
