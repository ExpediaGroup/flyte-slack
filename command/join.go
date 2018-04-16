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

package command

import (
	"encoding/json"
	"fmt"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/flyte-slack/client"
)

var (
	channelJoinedEventDef     = flyte.EventDef{Name: "ChannelJoined"}
	joinChannelFailedEventDef = flyte.EventDef{Name: "JoinChannelFailed"}
)

type JoinChannelInput struct {
	ChannelId string `json:"channelId"`
}

type JoinChannelOutput struct {
	JoinChannelInput
}

type JoinChannelFailedOutput struct {
	JoinChannelOutput
	Error string `json:"error"`
}

func JoinChannel(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "JoinChannel",
		OutputEvents: []flyte.EventDef{channelJoinedEventDef, joinChannelFailedEventDef},
		Handler:      joinChannelHandler(slack),
	}
}

func joinChannelHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {

		input := JoinChannelInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("input is not valid: %v", err))
		}

		if input.ChannelId == "" {
			return newJoinChannelFailedEvent(input.ChannelId, "missing channel id field")
		}
		slack.JoinChannel(input.ChannelId)
		return newChannelJoinedEvent(input.ChannelId)
	}
}

func newChannelJoinedEvent(channelId string) flyte.Event {

	return flyte.Event{
		EventDef: channelJoinedEventDef,
		Payload:  JoinChannelOutput{JoinChannelInput: JoinChannelInput{ChannelId: channelId}},
	}
}

func newJoinChannelFailedEvent(channelId string, err string) flyte.Event {

	output := JoinChannelOutput{JoinChannelInput: JoinChannelInput{ChannelId: channelId}}
	return flyte.Event{
		EventDef: joinChannelFailedEventDef,
		Payload:  JoinChannelFailedOutput{JoinChannelOutput: output, Error: err},
	}
}
