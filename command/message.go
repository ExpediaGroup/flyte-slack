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
	"github.com/ExpediaGroup/flyte-client/flyte"
	"github.com/ExpediaGroup/flyte-slack/client"
	"strings"
)

var (
	messageSentEventDef       = flyte.EventDef{Name: "MessageSent"}
	sendMessageFailedEventDef = flyte.EventDef{Name: "SendMessageFailed"}
)

type SendMessageInput struct {
	Message         string `json:"message"`
	ThreadTimestamp string `json:"threadTimestamp"`
	ChannelId       string `json:"channelId"`
}

type SendMessageOutput struct {
	SendMessageInput
}

type SendMessageErrorOutput struct {
	SendMessageOutput
	Error string `json:"error"`
}

func SendMessage(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "SendMessage",
		OutputEvents: []flyte.EventDef{messageSentEventDef, sendMessageFailedEventDef},
		Handler:      sendMessageHandler(slack),
	}
}

func sendMessageHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {

		input := SendMessageInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("input is not valid: %v", err))
		}

		errorMessages := []string{}
		if input.Message == "" {
			errorMessages = append(errorMessages, "missing message field")
		}
		if input.ChannelId == "" {
			errorMessages = append(errorMessages, "missing channel id field")
		}
		if len(errorMessages) != 0 {
			return newSendMessageFailedEvent(input.Message, input.ChannelId, strings.Join(errorMessages, ", "))
		}

		slack.SendMessage(input.Message, input.ChannelId, input.ThreadTimestamp)
		return newMessageSentEvent(input.Message, input.ChannelId)
	}
}

func newMessageSentEvent(message, channelId string) flyte.Event {

	return flyte.Event{
		EventDef: messageSentEventDef,
		Payload:  SendMessageOutput{SendMessageInput: SendMessageInput{Message: message, ChannelId: channelId}},
	}
}

func newSendMessageFailedEvent(message, channelId string, err string) flyte.Event {

	output := SendMessageOutput{SendMessageInput: SendMessageInput{Message: message, ChannelId: channelId}}
	return flyte.Event{
		EventDef: sendMessageFailedEventDef,
		Payload:  SendMessageErrorOutput{SendMessageOutput: output, Error: err},
	}
}
