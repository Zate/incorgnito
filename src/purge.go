package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/diamondburned/arikawa/bot"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

const (
	clf = "ChannelList.json"
	hrs = 72
)

// Purger struct
type Purger struct {
	ticker *time.Ticker                            // periodic ticker
	add    chan discord.ChannelID                  // new channelID
	remove chan discord.ChannelID                  // remove channelID
	delmsg chan discord.Message                    // Message to delete
	chids  []discord.ChannelID                     // current channelIDs to purge
	last   map[discord.ChannelID]discord.MessageID // last message ID processed per channel
}

// ChannelList struct to keep track of the list of item
type ChannelList struct {
	Channels []discord.ChannelID `json:"channels,omitempty"`
}

// NewPurger returns a new *Purger
func NewPurger() *Purger {
	return &Purger{
		ticker: time.NewTicker(time.Second * 30),
		add:    make(chan discord.ChannelID),
		remove: make(chan discord.ChannelID),
		delmsg: make(chan discord.Message, 20),
		chids:  []discord.ChannelID{},
		last:   make(map[discord.ChannelID]discord.MessageID),
	}
}

// PurgePackage is a bunch of stuff to purge
type PurgePackage struct {
	hrs    int
	now    time.Time
	chInfo *discord.Channel
	gInfo  *discord.Guild
	msgs   []discord.Message
	ctx    bot.Context
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
				if result != "" {
					log.Println(result)
				}
			}
			took := time.Now().Sub(start)
			log.Println("( Global ) [ Purger Cycle Completed in", fmt.Sprint(took.Round(took).Truncate(took).Milliseconds()), "ms ]")
		case u := <-p.add:
			re, _ := itemExists(p.chids, u)
			if re == false {
				p.chids = append(p.chids, u)
				log.Println("(", bot.gInfo(u).Name, ") [", u, "added to Purger Channel List ]")
				err := bot.P.saveChannelList()
				if err != nil {
					log.Println("Error saving ChannelList.json", err)
				}
				break
			}
			log.Println("(", bot.gInfo(u).Name, ") [", u, "already in Purger Channel List ]")
		case r := <-p.remove:
			re, i := itemExists(p.chids, r)
			if re == true {
				p.chids = append(p.chids[:i], p.chids[i+1:]...)
				log.Println("(", bot.gInfo(r).Name, ") [", r, "removed from Purger ]")
				err := bot.P.saveChannelList()
				if err != nil {
					log.Println("Error saving ChannelList.json", err)
				}
				break
			}
			log.Println("(", bot.gInfo(r).Name, ") [", r, "not found in Purger Channel List ]")
		case d := <-p.delmsg:
			log.Println("Got New Msg To Delete:", d.ID, "in", bot.cInfo(d.ChannelID).Name)
			err := bot.Ctx.DeleteMessage(d.ChannelID, d.ID)
			if err != nil {
				log.Println("Error Deleting", d.ID, "in", bot.cInfo(d.ChannelID).Name, "Error:", err)
			}
		}
	}
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

// MsgPurge purges messages in a Channel older than 72 hours.
// Todo: Make the 72 hours configurable via '!purge on <hrs>'
func (bot *Bot) MsgPurge(c discord.ChannelID) (string, error) {
	// log.Println("Last:", bot.P.last[c])
	count := 1
	pkg := PurgePackage{}
	pkg.hrs = hrs
	pkg.now = time.Now()
	pkg.chInfo, _ = bot.Ctx.Channel(c)
	pkg.gInfo, _ = bot.Ctx.Guild(pkg.chInfo.GuildID)
	// if bot.P.last[c] == 0 {
	bot.P.last[c] = pkg.chInfo.LastMessageID
	// }
	// log.Println("Last:", bot.P.last[c])

	for {
		msgs, err := bot.Ctx.MessagesBefore(pkg.chInfo.ID, bot.P.last[c], 100)
		if err != nil {
			text := "Failed getting Messages: " + fmt.Sprint(err)
			return text, err
		}
		//log.Printf("msgs: %v", fmt.Sprint(msgs))
		pkg.msgs = msgs
		i := len(pkg.msgs)
		if i == 0 {
			break
		}
		t := 0
		log.Println("(", pkg.gInfo.Name, ") [ Got", len(pkg.msgs), "in", pkg.chInfo.Name, " FID:", pkg.msgs[0].ID, "LID:", pkg.msgs[i-1].ID, "]")
		for q, o := range pkg.msgs {
			et := pkg.now.Sub(o.Timestamp.Time()).Round(1 * time.Hour).Hours()
			if et > float64(pkg.hrs) {
				bot.P.delmsg <- o
				// err := bot.Ctx.DeleteMessage(o.ChannelID, o.ID)
				// if err != nil {
				// 	log.Println("Error Deleting", o.ID, "in", bot.cInfo(o.ChannelID).Name, "with et of", et, "and Error:", err)
				// }
				log.Printf("( %v ) [%v] [ Added %v to delete queue from %v ]", pkg.gInfo.Name, q, o.ID, pkg.chInfo.Name)
			}
			//	log.Printf("( %v ) [%v] [ Processed %v from %v ]", pkg.gInfo.Name, q, o.ID, pkg.chInfo.Name)
			t = q
			//log.Println(q, count, et)
			count++

		}
		if i != t+1 {
			log.Println("Error: We had:", i, " to be processed but only looped over:", t)
		}
		bot.P.last[c] = pkg.msgs[t].ID
		if i < 100 {
			break
		}

		//	log.Println("Enf of loop - new lmid:", bot.P.last[c])
	}
	log.Printf("( %v ) [ Completed %v from %v ]", pkg.gInfo.Name, count-1, pkg.chInfo.Name)
	return "", nil
}

func (p *Purger) loadChannelList() error {
	if doExist(clf) == false {
		_, err := os.Create(clf)
		if err != nil {
			log.Println("Failed to Create:", err)
			return err
		}
	}
	fi, err := os.Stat(clf)
	if err != nil {
		log.Println("Failed to Stat:", err)
		return err
	}
	if fi.Size() == 0 {
		cl := ChannelList{}
		clout, err := json.MarshalIndent(cl.Channels, "", " ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(clf, clout, 0775)
		if err != nil {
			return err
		}
		return nil
	}
	f, err := ioutil.ReadFile(clf)
	if err != nil {
		log.Println("Failed To Read:", err)
		return err
	}

	err = json.Unmarshal([]byte(f), &p.chids)
	if err != nil {
		log.Println("Failed to Write:", err)
		return err
	}
	return nil
}

func (p *Purger) saveChannelList() error {
	cl := ChannelList{}
	cl.Channels = p.chids
	clout, err := json.MarshalIndent(cl.Channels, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(clf, clout, 0644)
	if err != nil {
		return err
	}
	return nil
}
