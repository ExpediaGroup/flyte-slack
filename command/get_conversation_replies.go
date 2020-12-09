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
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/slack-go/slack"
	"strings"
)

var (
	getConversationRepliesEventDef       = flyte.EventDef{Name: "GetConversationReplies"}
	getConversationRepliesFailedEventDef = flyte.EventDef{Name: "GetConversationRepliesFailed"}
)

type GetConversationRepliesInput struct {
	ThreadTimestamp string `json:"threadTimestamp"`
	ChannelId       string `json:"channelId"`
}

type GetConversationRepliesOutput struct {
	Message []slack.Message `json:"messages"`
}

type GetConversationRepliesErrorOutput struct {
	GetConversationRepliesOutput
	Error string `json:"error"`
}

func GetConversationReplies(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "GetConversationReplies",
		OutputEvents: []flyte.EventDef{getConversationRepliesEventDef, getConversationRepliesFailedEventDef},
		Handler:      getConversationRepliesHandler(slack),
	}
}

func getConversationRepliesHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {

		input := GetConversationRepliesInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("invalid input: %v", err))
		}

		errorMessages := []string{}
		if input.ThreadTimestamp == "" {
			errorMessages = append(errorMessages, "missing threadTimestamp field")
		}
		if input.ChannelId == "" {
			errorMessages = append(errorMessages, "missing channel id field")
		}
		if len(errorMessages) != 0 {
			return newGetConversationRepliesFailedEvent(nil, strings.Join(errorMessages, ", "))
		}

		slackReplies, err := slack.GetConversationReplies(input.ChannelId, input.ThreadTimestamp)

		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("there was an error retrieving the channel slackReplies: %v", err))
		}

		return newGetConversationRepliesEvent(slackReplies)
	}
}

func newGetConversationRepliesEvent(message []slack.Message) flyte.Event {

	return flyte.Event{
		EventDef: getConversationRepliesEventDef,
		Payload:  GetConversationRepliesOutput{Message: message},
	}
}

func newGetConversationRepliesFailedEvent(message []slack.Message, err string) flyte.Event {

	output := GetConversationRepliesOutput{Message: message}
	return flyte.Event{
		EventDef: getConversationRepliesFailedEventDef,
		Payload:  GetConversationRepliesErrorOutput{GetConversationRepliesOutput: output, Error: err},
	}
}
