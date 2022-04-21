/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/ExpediaGroup/flyte-client/flyte"
	"github.com/ExpediaGroup/flyte-slack/cache"
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/ExpediaGroup/flyte-slack/command"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/url"
	"time"
)

func main() {
	zerolog.SetGlobalLevel(logLevel())

	slack := client.NewSlack(slackToken())
	cc, err := cacheConfig()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	cache := cache.New(cc)

	pack := flyte.NewPackWithPolling(packDef(slack, cache), 1*time.Second)
	pack.Start()

	for e := range slack.IncomingMessages() {
		pack.SendEvent(e)
	}
}

func packDef(slack client.Slack, cache cache.Cache) flyte.PackDef {
	helpUrl, _ := url.Parse("https://github.com/ExpediaGroup/flyte-slack/blob/master/README.md")

	return flyte.PackDef{
		Name:    packName(),
		HelpURL: helpUrl,
		Commands: []flyte.Command{
			command.SendMessage(slack),
			command.SendRichMessage(slack),
			command.GetChannelInfo(slack, cache),
		},
		EventDefs: []flyte.EventDef{
			{Name: "ReceivedMessage"},
		},
	}
}
