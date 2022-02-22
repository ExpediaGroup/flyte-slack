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
	"github.com/ExpediaGroup/flyte-slack/cache"
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/ExpediaGroup/flyte-slack/command"
	api "github.com/HotelsDotCom/flyte-client/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
	"net/url"
	"time"
)

const packDefHelpUrl = "https://github.com/ExpediaGroup/flyte-slack/blob/master/README.md"
const defaultPackName = "Slack"

func main() {

	slack := client.NewSlack(slackToken())
	cc, err := cacheConfig()
	if err != nil {
		logger.Fatal(err)
	}

	caches := cache.New(cc)
	packDef := GetPackDef(slack, caches)
	pack := flyte.NewPack(packDef, api.NewClient(apiHost(), 10*time.Second))
	pack.Start()

	ListenAndServe(slack, pack)

}

func ListenAndServe(slack client.Slack, pack flyte.Pack) {

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)

	// handle incoming messages
	incomingMessages := slack.IncomingMessages()
	go func() {
		for e := range incomingMessages {
			pack.SendEvent(e)
		}
	}()

	select {}
}

func GetPackDef(slack client.Slack, cache cache.Cache) flyte.PackDef {

	helpUrl, err := url.Parse(packDefHelpUrl)
	if err != nil {
		logger.Fatal("invalid pack help url")
	}

	packName := packName()
	if packName == "" {
		packName = defaultPackName
	}

	return flyte.PackDef{
		Name:    packName,
		HelpURL: helpUrl,
		Commands: []flyte.Command{
			command.SendMessage(slack),
			command.SendRichMessage(slack),
			command.GetChannelInfo(slack, cache),
			command.GetReactionList(slack),
		},
		EventDefs: []flyte.EventDef{
			{Name: "ReceivedMessage"},
			{Name: "ReactionAdded"},
		},
	}
}
