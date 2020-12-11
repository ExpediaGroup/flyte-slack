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
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConversationRepliesCommandIsPopulated(t *testing.T) {
	command := CreateConversationReplies(MessageMockSlack)

	assert.Equal(t, "CreateConversationReplies", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "GetConversationRepliesSuccess", command.OutputEvents[0].Name)
	assert.Equal(t, "GetConversationRepliesFailed", command.OutputEvents[1].Name)
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithInvalidJSON(t *testing.T) {
	cmd := CreateConversationReplies(nil)

	event := cmd.Handler([]byte(`.`))

	fatalEventDef := flyte.NewFatalEvent("").EventDef
	assert.Equal(t, fatalEventDef, event.EventDef)
	assert.Contains(t, event.Payload.(string), "invalid input: ")
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithMissingTimestamp(t *testing.T) {
	BeforeMessage()
	defer AfterMessage()

	handler := CreateConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"channelId": "UXB456Y"}`))

	output := event.Payload.(GetConversationRepliesErrorOutput)
	assert.Equal(t, "GetConversationRepliesFailed", event.EventDef.Name)
	assert.Equal(t, "missing threadTimestamp field", output.Error)
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithMissingChannelId(t *testing.T) {
	BeforeMessage()
	defer AfterMessage()

	handler := CreateConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"threadTimestamp": "333333"}`))

	output := event.Payload.(GetConversationRepliesErrorOutput)
	assert.Equal(t, "GetConversationRepliesFailed", event.EventDef.Name)
	assert.Equal(t, "missing channel id field", output.Error)
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithMissingChannelIdAndTimestamp(t *testing.T) {
	BeforeMessage()
	defer AfterMessage()

	handler := CreateConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"abc": "123"}`))

	output := event.Payload.(GetConversationRepliesErrorOutput)
	assert.Equal(t, "GetConversationRepliesFailed", event.EventDef.Name)
	assert.Equal(t, "missing threadTimestamp field", output.Error)
}

func TestGetConversationRepliesReturnsGetConversationRepliesEvent(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	// Mirror what is on https://api.slack.com/messaging/retrieving#conversations
	slackReplies := []slack.Message{
		{
			Msg: slack.Msg{
				ReplyCount:      3,
				Type:            "message",
				User:            "Greg",
				ThreadTimestamp: "1234568780",
				Replies: []slack.Reply{
					{
						Timestamp: "123",
						User:      "Karl",
					},
				},
			},
		},
		{
			Msg: slack.Msg{
				Type:            "message",
				User:            "Tom",
				Text:            "abc",
				ThreadTimestamp: "1234",
				ParentUserId:    "1234",
			},
		},
	}

	MessageMockSlack.GetConversationRepliesFunc = func(channel string, threadTimestamp string) ([]slack.Message, error) {
		return slackReplies, nil
	}

	handler := CreateConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"threadTimestamp": "yo", "channelId": "xyz"}`))

	output := event.Payload.(GetConversationRepliesOutput)
	require.Equal(t, "1234568780", output.Messages[0].ThreadTimestamp)
	require.Equal(t, "1234", output.Messages[1].ThreadTimestamp)
	require.Equal(t, "Greg", output.Messages[0].User)
	require.Equal(t, "Tom", output.Messages[1].User)
}
