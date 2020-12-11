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
	"errors"
	"fmt"
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/slack-go/slack"
)

var (
	getConversationRepliesEventDef       = flyte.EventDef{Name: "GetConversationRepliesSuccess"}
	getConversationRepliesFailedEventDef = flyte.EventDef{Name: "GetConversationRepliesFailed"}
)

type GetConversationRepliesInput struct {
	ThreadTimestamp string `json:"threadTimestamp"`
	ChannelId       string `json:"channelId"`
}

type GetConversationRepliesOutput struct {
	Messages []slack.Message `json:"messages"`
}

type GetConversationRepliesErrorOutput struct {
	Error string `json:"error"`
}

func CreateConversationReplies(slack client.Slack) flyte.Command {
	return flyte.Command{
		Name: "CreateConversationReplies",
		OutputEvents: []flyte.EventDef{
			getConversationRepliesEventDef,
			getConversationRepliesFailedEventDef,
		},
		Handler: createHandler(slack),
	}
}

func createHandler(slack client.Slack) func(json.RawMessage) flyte.Event {
	return func(rawInput json.RawMessage) flyte.Event {

		input := GetConversationRepliesInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("invalid input: %v", err))
		}

		if err := validateInput(input); err != nil {
			return newGetConversationRepliesFailedEvent(err.Error())
		}

		slackReplies, err := slack.GetConversationReplies(input.ChannelId, input.ThreadTimestamp)

		if err != nil {
			return newGetConversationRepliesFailedEvent(fmt.Errorf("there was an error retrieving the channel slackReplies: %v", err).Error())
		}

		return newGetConversationRepliesEvent(slackReplies)
	}
}

func validateInput(in GetConversationRepliesInput) error {
	if in.ThreadTimestamp == "" {
		return errors.New("missing threadTimestamp field")
	}
	if in.ChannelId == "" {
		return errors.New("missing channel id field")
	}
	return nil
}

func newGetConversationRepliesEvent(message []slack.Message) flyte.Event {
	return flyte.Event{
		EventDef: getConversationRepliesEventDef,
		Payload:  GetConversationRepliesOutput{Messages: message},
	}
}

func newGetConversationRepliesFailedEvent(err string) flyte.Event {
	return flyte.Event{
		EventDef: getConversationRepliesFailedEventDef,
		Payload:  GetConversationRepliesErrorOutput{Error: err},
	}
}
