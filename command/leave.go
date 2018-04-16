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
	channelLeftEventDef        = flyte.EventDef{Name: "ChannelLeft"}
	leaveChannelFailedEventDef = flyte.EventDef{Name: "LeaveChannelFailed"}
)

type LeaveChannelInput struct {
	ChannelId string `json:"channelId"`
}

type LeaveChannelOutput struct {
	LeaveChannelInput
}

type LeaveChannelFailedOutput struct {
	LeaveChannelOutput
	Error string `json:"error"`
}

func LeaveChannel(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "LeaveChannel",
		OutputEvents: []flyte.EventDef{channelLeftEventDef, leaveChannelFailedEventDef},
		Handler:      leaveChannelHandler(slack),
	}
}

func leaveChannelHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {

		input := LeaveChannelInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("input is not valid: %v", err))
		}

		if input.ChannelId == "" {
			return newLeaveChannelFailedEvent(input.ChannelId, "missing channel id field")
		}
		slack.LeaveChannel(input.ChannelId)
		return newChannelLeftEvent(input.ChannelId)
	}
}

func newChannelLeftEvent(channelId string) flyte.Event {

	return flyte.Event{
		EventDef: channelLeftEventDef,
		Payload:  LeaveChannelOutput{LeaveChannelInput: LeaveChannelInput{ChannelId: channelId}},
	}
}

func newLeaveChannelFailedEvent(channelId string, err string) flyte.Event {

	output := LeaveChannelOutput{LeaveChannelInput: LeaveChannelInput{ChannelId: channelId}}
	return flyte.Event{
		EventDef: leaveChannelFailedEventDef,
		Payload:  LeaveChannelFailedOutput{LeaveChannelOutput: output, Error: err},
	}
}
