package zenbot

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aybabtme/log"
	"github.com/nlopes/slack"
)

// Slack contains the Slack client and RTM object.
type Slack struct {
	Bot *slack.Client
	RTM *slack.RTM
}

// Config contains all the necessary configs for a
// zenbot instance.
type Config struct {
	Slack            *Slack
	Debug            bool
	Log              *log.Log
	TimeoutDuration  time.Duration
	ChannelWhitelist StringList
}

// A Zen is a zen time period for a user
type Zen struct {
	User, Name, Channel, Reason string
	EndsAt, Timeout             time.Time
}

// A Bot is an instance of zenbot.
type Bot struct {
	Config       *Config
	zens         []*Zen
	zensMutex    sync.RWMutex
	channelNames map[string]string
}

var regexps = struct {
	Zen, ZenArgs, Cancel *regexp.Regexp
}{
	Zen:     regexp.MustCompile(`^\.\/zen`),
	ZenArgs: regexp.MustCompile(`^\.\/zen +t?((?:\d+h)?(?:\d+m)?(?:\d+s)?)(?: (.*)?)$`),
	Cancel:  regexp.MustCompile(`^\.\/zen cancel(?: (.*))?$`),
}

func New(config *Config) *Bot {
	return &Bot{
		Config:       config,
		channelNames: make(map[string]string),
	}
}

// Zen starts listening for Slack messages.
func (b *Bot) Zen() {
	go b.ExpireZens()

	for {
		select {
		case msg := <-b.Config.Slack.RTM.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				go b.handleMessageEvent(msg.Data.(*slack.MessageEvent))
			case *slack.ConnectedEvent:
				b.Config.Log.Info("connected to slack")

				if b.Config.Debug {
					b.Config.Log.KV("info", ev.Info).Info("got slack info")
					b.Config.Log.KV("connections", ev.ConnectionCount).Info("got connection count")
				}
			case *slack.RTMError:
				b.Config.Log.Err(ev).Error("slack rtm error")
			case *slack.InvalidAuthEvent:
				b.Config.Log.Fatal("invalid slack token")
			// user activity events
			case *slack.UserTypingEvent:
				go b.enforceZen(ev.User, "typing")
			case *slack.ReactionAddedEvent:
				go b.enforceZen(ev.User, "using reactjis")
			case *slack.ReactionRemovedEvent:
				go b.enforceZen(ev.User, "using reactjis")
			case *slack.StarRemovedEvent:
				go b.enforceZen(ev.User, "starring messages")
			case *slack.StarAddedEvent:
				go b.enforceZen(ev.User, "starring messages")
			case *slack.PinRemovedEvent:
				go b.enforceZen(ev.User, "pinning messages")
			case *slack.PinAddedEvent:
				go b.enforceZen(ev.User, "pinning messages")
			default:
			}
		}
	}
}

// SendMessage sends a message to a Slack channel.
func (b *Bot) SendMessage(message, channel string) {
	b.Config.Slack.RTM.SendMessage(b.Config.Slack.RTM.NewOutgoingMessage(message, channel))
}

func (b *Bot) handleError(err error, channel string) bool {
	if err == nil {
		return false
	}

	b.Config.Log.Err(err).Error("error")

	b.SendMessage(err.Error(), channel)
	return true
}

func (b *Bot) handleMessageEvent(ev *slack.MessageEvent) {
	if ev.Type != "message" {
		return
	}

	switch {
	case regexps.Cancel.MatchString(ev.Text):
		if b.isWhitelisted(ev) {
			b.cancelZen(ev)
		}
	case regexps.Zen.MatchString(ev.Text):
		if b.isWhitelisted(ev) {
			b.startZen(ev)
		}
	default:
	}
}

func (b *Bot) isWhitelisted(ev *slack.MessageEvent) bool {
	if len(b.Config.ChannelWhitelist) == 0 {
		return true
	}

	name, err := b.getChannelName(ev.Channel)
	if err != nil {
		b.Config.Log.KV("id", ev.Channel).Err(err).Error("could not look up channel name")
		return false
	}

	_, whitelisted := b.Config.ChannelWhitelist[name]
	if b.Config.Debug {
		b.Config.Log.KV("channel", name).Info("channel is not whitelisted, command ignored")
	}
	return whitelisted
}

func (b *Bot) startZen(ev *slack.MessageEvent) {
	match := regexps.ZenArgs.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		b.SendMessage("Usage: `./zen <duration e.g. 1h30m> [reason - optional]`", ev.Channel)
		return
	}

	durationString, reason := match[1], match[2]
	duration, err := time.ParseDuration(durationString)

	if b.handleError(err, ev.Channel) {
		return
	}

	name, err := b.getUserName(ev.User)
	if b.handleError(err, ev.Channel) {
		return
	}

	zen := &Zen{
		User:    ev.User,
		Name:    name,
		Channel: ev.Channel,
		Reason:  reason,
		EndsAt:  time.Now().Add(duration),
		Timeout: time.Now().Add(5 * time.Second),
	}

	b.zensMutex.Lock()
	b.zens = append(b.zens, zen)
	b.zensMutex.Unlock()

	b.SendMessage(fmt.Sprintf("Added a zen for %s (%s), ends at [%s].", durationString, reason, zen.EndsAt), ev.Channel)
}

func (b *Bot) cancelZen(ev *slack.MessageEvent) {
	match := regexps.Cancel.FindStringSubmatch(ev.Text)
	if len(match) == 0 {
		b.SendMessage("Usage: `./zen cancel [reason - optional]`", ev.Channel)
		return
	}

	var (
		reason  = match[1]
		message = ""
		count   = 0
	)

	b.zensMutex.RLock()
	for i := 0; i < len(b.zens); i++ {
		zen := b.zens[i]
		if zen.User == ev.User && (reason == "" || strings.ToLower(zen.Reason) == strings.ToLower(reason)) {
			b.zensMutex.RUnlock()

			b.zensMutex.Lock()
			b.zens = append(b.zens[:i], b.zens[i+1:]...)
			i--
			b.zensMutex.Unlock()

			message += fmt.Sprintf("(%s) zen canceled.\n", zen.Reason)
			count++

			b.zensMutex.RLock()
		}
	}
	b.zensMutex.RUnlock()

	if message == "" {
		if reason != "" {
			message = "you do not have any such running zens"
		} else {
			message = "you do not have any running zens"
		}
	}
	b.SendMessage(message, ev.Channel)

	if count > 0 {
		b.SendMessage(fmt.Sprintf("<@%s>-%s for canceling %d zens.", ev.User, strings.Repeat("-", count), count), ev.Channel)
	}

}

func (b *Bot) enforceZen(user, activity string) {
	b.zensMutex.RLock()
	defer b.zensMutex.RUnlock()

	for i := 0; i < len(b.zens); i++ {
		zen := b.zens[i]
		if zen.User == user && time.Now().After(zen.Timeout) {
			b.SendMessage(fmt.Sprintf("<@%s>-- for %s during your zen period (%s).", zen.User, activity, zen.Reason), zen.Channel)

			b.zensMutex.RUnlock()
			b.zensMutex.Lock()
			zen.Timeout = time.Now().Add(b.Config.TimeoutDuration)
			b.zensMutex.Unlock()
			b.zensMutex.RLock()
			break
		}
	}
}

// ExpireZens removes zens that have ended
func (b *Bot) ExpireZens() {
	var wg sync.WaitGroup

	for {
		wg.Wait()
		<-time.After(1 * time.Second)

		wg.Add(1)
		go func() {
			b.zensMutex.RLock()

			now := time.Now()
			for i, zen := range b.zens {
				if now.After(zen.EndsAt) {
					b.zensMutex.RUnlock()

					b.zensMutex.Lock()
					b.zens = append(b.zens[:i], b.zens[i+1:]...)
					b.zensMutex.Unlock()

					b.SendMessage(fmt.Sprintf("<@%s>: Be free, for your zen (%s) has ended!", zen.User, zen.Reason), zen.Channel)

					b.zensMutex.RLock()
				}
			}

			b.zensMutex.RUnlock()
			wg.Done()
		}()
	}
}

func (b *Bot) getUserName(id string) (string, error) {
	userInfo, err := b.Config.Slack.Bot.GetUserInfo(id)
	if err != nil {
		return "", err
	}

	return userInfo.Name, nil
}

func (b *Bot) getChannelName(id string) (string, error) {
	if name, ok := b.channelNames[id]; ok {
		return name, nil
	}

	channelInfo, err := b.Config.Slack.Bot.GetChannelInfo(id)
	if err != nil {
		return "", err
	}

	b.channelNames[id] = channelInfo.Name
	return channelInfo.Name, nil
}
