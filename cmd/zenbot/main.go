package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/kamaln7/zenbot"

	"github.com/aybabtme/log"
	"github.com/nlopes/slack"
)

var (
	token            = flag.String("token", "", "slack RTM token")
	debug            = flag.Bool("debug", false, "toggle debug")
	timeout          = flag.String("timeout", "10s", "timeout between karma operations")
	channelWhitelist = make(zenbot.StringList, 0)
)

func main() {
	log.Info(fmt.Sprintf("starting zenbot %s", zenbot.Version))

	flag.Var(&channelWhitelist, "whitelist.chan", "set a list of channels that zenbot may be used in")
	flag.Parse()
	timeoutDuration, err := time.ParseDuration(*timeout)
	if err != nil {
		log.KV("timeout", *timeout).Err(err).Fatal("could not parse timeout duration")
	}

	sc := &zenbot.Slack{
		Bot: slack.New(*token),
	}
	sc.Bot.SetDebug(*debug)
	sc.RTM = sc.Bot.NewRTM()

	go sc.RTM.ManageConnection()

	bot := zenbot.New(&zenbot.Config{
		Slack:            sc,
		Log:              log.KV("zenbot", "true"),
		Debug:            *debug,
		TimeoutDuration:  timeoutDuration,
		ChannelWhitelist: channelWhitelist,
	})

	bot.Zen()
}
