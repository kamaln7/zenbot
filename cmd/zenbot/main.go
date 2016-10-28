package main

import (
	"flag"

	"github.com/kamaln7/zenbot"

	"github.com/aybabtme/log"
	"github.com/nlopes/slack"
)

var (
	token = flag.String("token", "", "slack RTM token")
	debug = flag.Bool("debug", false, "toggle debug")
)

func main() {
	log.Info("starting zenbot")

	flag.Parse()

	sc := &zenbot.Slack{
		Bot: slack.New(*token),
	}
	sc.Bot.SetDebug(*debug)
	sc.RTM = sc.Bot.NewRTM()

	go sc.RTM.ManageConnection()

	bot := &zenbot.Bot{
		Config: &zenbot.Config{
			Slack: sc,
			Log:   log.KV("zenbot", "true"),
			Debug: *debug,
		},
	}

	bot.Zen()
}
