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
	"net/url"
	"os"
	"os/signal"
	api "github.com/HotelsDotCom/flyte-client/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/flyte-slack/client"
	"github.com/HotelsDotCom/flyte-slack/command"
	"github.com/HotelsDotCom/go-logger"
	"syscall"
	"time"
)

const packDefHelpUrl = "https://github.com/HotelsDotCom/flyte-slack/blob/master/README.md"

func main() {

	slack := client.NewSlack(Backup(), SlackToken(), DefaultChannel())
	packDef := GetPackDef(slack)
	pack := flyte.NewPack(packDef, api.NewClient(ApiHost(), 10*time.Second))
	pack.Start()

	ListenAndServe(slack, pack)
}

func Backup() client.Backup {

	backupDir := BackupDir()
	if backupDir == "" {
		return client.NewTmpFileBackup()
	}
	return client.NewFileBackup(backupDir)
}

func ListenAndServe(slack client.Slack, pack flyte.Pack) {

	// handle incoming messages
	incomingMessages := slack.IncomingMessages()
	go func() {
		for e := range incomingMessages {
			pack.SendEvent(e)
		}
	}()

	// block until we get an exit causing signal
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		logger.Info("received interrupt, shutting down ...")
		slack.Broadcast("slack shutting down ...")
		time.Sleep(2 * time.Second)
		logger.Info("shut down")
	}
}

func GetPackDef(slack client.Slack) flyte.PackDef {

	helpUrl, err := url.Parse(packDefHelpUrl)
	if err != nil {
		logger.Fatal("invalid pack help url")
	}

	return flyte.PackDef{
		Name:    "Slack",
		HelpURL: helpUrl,
		Commands: []flyte.Command{
			command.SendMessage(slack),
			command.SendRichMessage(slack),
			command.Broadcast(slack),
			command.JoinChannel(slack),
			command.LeaveChannel(slack),
		},
		EventDefs: []flyte.EventDef{
			{Name: "ReceivedMessage"},
		},
	}
}
