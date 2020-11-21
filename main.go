package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/diamondburned/arikawa/bot"
	"github.com/diamondburned/arikawa/bot/extras/middlewares"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
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
		me, err := ctx.Me()
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("( Initializing ... ) [ %v is alive ]", me.Username)
		guilds, err := ctx.Guilds()
		if err != nil {
			log.Fatalln(err)
		}
		for _, g := range guilds {
			log.Printf("( %v ) [ %v bot joined server %v ]", g.Name, me.Username, g.Name)
		}
		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	if err := wait(); err != nil {
		log.Fatalln("Gateway fatal error:", err)
	}
}

// Bot struct
type Bot struct {
	// Context must not be embedded.
	Ctx *bot.Context
	P   *Purger
}

// Setup bot
func (bot *Bot) Setup(sub *bot.Subcommand) {
	// Only allow people in guilds to run guildInfo.
	sub.AddMiddleware("GuildInfo", middlewares.GuildOnly(bot.Ctx))
	sub.AddMiddleware("Purge", middlewares.AdminOnly(bot.Ctx))
	bot.P = NewPurger()
	go bot.runPurger(bot.P)
}

// Purge {on|off}lets an Admin add this channel to the Purger Channel List
func (bot *Bot) Purge(m *gateway.MessageCreateEvent, l ...string) (string, error) {
	if len(l) == 0 {
		return bot.Ctx.Help(), nil
	}

	chinfo, _ := bot.Ctx.Channel(m.ChannelID)
	re, _ := itemExists(bot.P.chids, m.ChannelID)

	if l[0] == "on" {
		if re == false {
			bot.P.add <- m.ChannelID
			text := fmt.Sprint(chinfo.Name) + " added to Purger Channel List"
			return text, nil
		}
		text := fmt.Sprint(chinfo.Name) + " already on Purger Channel List"
		return text, nil
	}
	if l[0] == "off" {
		if re == true {
			bot.P.remove <- m.ChannelID
			text := fmt.Sprint(chinfo.Name) + " removed from Purger Channel List"
			return text, nil
		}
		text := fmt.Sprint(chinfo.Name) + " is not in Purger Channel List"
		return text, nil
	}
	return bot.Ctx.Help(), nil
}

// Help prints the default help message.
func (bot *Bot) Help(*gateway.MessageCreateEvent) (string, error) {
	return bot.Ctx.Help(), nil
}

// Ping is a simple ping example, perhaps the most simple you could make it.
func (bot *Bot) Ping(*gateway.MessageCreateEvent) (string, error) {
	return "Pong!", nil
}

// GuildInfo demonstrates the GuildOnly middleware done in (*Bot).Setup().
func (bot *Bot) GuildInfo(m *gateway.MessageCreateEvent) (string, error) {
	g, err := bot.Ctx.GuildWithCount(m.GuildID)
	if err != nil {
		return "", fmt.Errorf("failed to get guild: %v", err)
	}

	return fmt.Sprintf(
		"Your guild is %s, and its maximum members is %d",
		g.Name, g.ApproximateMembers,
	), nil
}

// MsgPurge purges messages in a Channel older than 72 hours.
// Todo: Make the 72 hours configurable via '!purge on <hrs>'
func (bot *Bot) MsgPurge(c discord.ChannelID) (string, error) {
	hrs := 72
	now := time.Now()
	chInfo, _ := bot.Ctx.Channel(c)
	gInfo, err := bot.Ctx.Guild(chInfo.GuildID)
	//var msgs []discord.Message
	// using bot.Ctx.Session forces us to go against the API
	// API has a max of 100 messages per attempt.
	// We should use MessagesBefore the last messageID
	// if that result is less than 100, we continue
	// if it's 100, we should take the last [100] message
	// and call again.  We need to append the results to msgs
	// continue this until the result from the API is les than 100.
	msgs, err := bot.Ctx.MessagesBefore(c, chInfo.LastMessageID, 100)
	if err != nil {
		text := "Failed getting Messages: " + fmt.Sprint(err)
		return text, err
	}
	if len(msgs) >= 100 {
		getMsgs := msgs
		lmid := msgs[99].ID
		log.Println(lmid, "message ID")
		for len(getMsgs) == 100 {
			ml := len(getMsgs) - 1
			lmid = getMsgs[ml].ID
			getMsgs, err := bot.Ctx.Session.MessagesBefore(c, lmid, 100)
			if err != nil {
				log.Printf("Failed getting Messages: %v", err)
				break
			}
			msgs = append(msgs, getMsgs...)
			if len(getMsgs) < 100 {
				break
			}
		}
	}

	var o discord.Message
	var count int
	var msgIDs []discord.MessageID
	log.Println("(", gInfo.Name, ") [ Processing", len(msgs), "in", chInfo.Name, "]")
	for _, o = range msgs {
		et := now.Sub(o.Timestamp.Time()).Round(1 * time.Hour).Hours()
		if et > float64(hrs) {
			count++
			if (now.Sub(o.Timestamp.Time()).Round(1 * time.Hour)).Hours() > float64(hrs) {
				_ = bot.Ctx.DeleteMessage(c, o.ID)
				log.Printf("(%v) [ Deleted ID: %v ]", gInfo.Name, o.ID)
			} else if count <= 99 {
				msgIDs = append(msgIDs, o.ID)
			}
		}
	}

	if len(msgIDs) > 2 {
		err = bot.Ctx.DeleteMessages(c, msgIDs)
		if err != nil {
			return fmt.Sprintf(
				"Something went wrong: %v",
				err,
			), err
		}
		log.Println("(", gInfo.Name, ") [ Deleted", len(msgIDs), "messages ]")
	}
	text := "( " + gInfo.Name + " ) [ Purged! ]"
	return text, nil
}

// Purger struct
type Purger struct {
	ticker *time.Ticker           // periodic ticker
	add    chan discord.ChannelID // new channelID
	remove chan discord.ChannelID // remove channelID
	chids  []discord.ChannelID    // current channelIDs to purge
}

// NewPurger returns a new *Purger
func NewPurger() *Purger {
	return &Purger{
		ticker: time.NewTicker(time.Second * 30),
		add:    make(chan discord.ChannelID),
		remove: make(chan discord.ChannelID),
		chids:  []discord.ChannelID{},
	}
}

// gInfo takes a ChannelID and reutns a Guild object from Ctx
func (bot *Bot) gInfo(c discord.ChannelID) *discord.Guild {
	var g *discord.Guild
	chInfo, _ := bot.Ctx.Channel(c)
	g, _ = bot.Ctx.Guild(chInfo.GuildID)
	return g
}

// runPurger takes a *Purger and montitors the channels and takes action when info comes in.
func (bot *Bot) runPurger(p *Purger) {
	for {
		select {
		case <-p.ticker.C:
			start := time.Now()
			log.Println("( Global ) [ Purger Cycle Started (30s) ]")
			log.Println("( Global ) [ Channels being purged:", len(p.chids), "]")
			for _, u := range p.chids {
				result, err := bot.MsgPurge(u)
				if err != nil {
					log.Println(err)
				}
				log.Println(result)
			}
			took := time.Now().Sub(start)
			log.Println("( Global ) [ Purger Cycle Completed in", fmt.Sprint(took.Round(took).Truncate(took).Milliseconds()), "ms ]")
		case u := <-p.add:
			re, _ := itemExists(p.chids, u)
			if re == false {
				p.chids = append(p.chids, u)
				log.Println("(", bot.gInfo(u).Name, ") [", u, "added to Purger Channel List ]")
				break
			}
			log.Println("(", bot.gInfo(u).Name, ") [", u, "already in Purger Channel List ]")
		case r := <-p.remove:
			re, i := itemExists(p.chids, r)
			if re == true {
				p.chids = append(p.chids[:i], p.chids[i+1:]...)
				log.Println("(", bot.gInfo(r).Name, ") [", r, "removed from Purger ]")
				break
			}
			log.Println("(", bot.gInfo(r).Name, ") [", r, "not found in Purger Channel List ]")
		}
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
