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
	broadcastSentEventDef   = flyte.EventDef{Name: "BroadcastSent"}
	broadcastFailedEventDef = flyte.EventDef{Name: "BroadcastFailed"}
)

type BroadcastInput struct {
	Message string `json:"message"`
}

type BroadcastOutput struct {
	BroadcastInput
}

type BroadcastErrorOutput struct {
	BroadcastOutput
	Error string `json:"error"`
}

func Broadcast(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "Broadcast",
		OutputEvents: []flyte.EventDef{broadcastSentEventDef, broadcastFailedEventDef},
		Handler:      broadcastHandler(slack),
	}
}

func broadcastHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {

		input := BroadcastInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("input is not valid: %v", err))
		}

		if input.Message == "" {
			return newBroadcastFailedEvent(input.Message, "missing message field")
		}

		slack.Broadcast(input.Message)
		return newBroadcastEvent(input.Message)
	}
}

func newBroadcastEvent(message string) flyte.Event {

	return flyte.Event{
		EventDef: broadcastSentEventDef,
		Payload:  BroadcastOutput{BroadcastInput: BroadcastInput{Message: message}},
	}
}

func newBroadcastFailedEvent(message string, err string) flyte.Event {

	output := BroadcastOutput{BroadcastInput: BroadcastInput{Message: message}}
	return flyte.Event{
		EventDef: broadcastFailedEventDef,
		Payload:  BroadcastErrorOutput{BroadcastOutput: output, Error: err},
	}
}
