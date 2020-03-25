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
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/flyte-slack/client"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPostMessageCommandIsPopulated(t *testing.T) {
	command := SendRichMessage(nil)

	assert.Equal(t, "SendRichMessage", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "RichMessageSent", command.OutputEvents[0].Name)
	assert.Equal(t, "SendRichMessageFailed", command.OutputEvents[1].Name)
}

func TestPostMessageShouldReturnFatalErrorEventWhenCalledWithInvalidJSON(t *testing.T) {
	cmd := SendRichMessage(nil)

	event := cmd.Handler([]byte(`.`))

	fatalEventDef := flyte.NewFatalEvent("").EventDef
	assert.Equal(t, fatalEventDef, event.EventDef)
	assert.Contains(t, event.Payload.(string), "invalid input: ")
}

func TestPostMessageShouldLogErrorWhenCalledWithInvalidJSON(t *testing.T) {
	loggertest.Init("DEBUG")

	cmd := SendRichMessage(nil)
	cmd.Handler([]byte(`.`))

	msgs := loggertest.GetLogMessages()
	require.Equal(t, 1, len(msgs))
	assert.Equal(t, loggertest.LogLevelError, msgs[0].Level)
	assert.True(t, strings.HasPrefix(msgs[0].Message, "invalid input: "), "expected message to start with \"invalid input: \", actual: %q", msgs[0].Message)

	loggertest.Reset()
}

func TestSendRichMessageHandlerShouldReturnErrorEventWhenRichMessageSenderReturnsError(t *testing.T) {
	mp := mockRichMessageSender{
		sendRichMessage: func(rm client.RichMessage) (string, string, error) {
			return "", "", errors.New("oh dear")
		},
	}

	command := SendRichMessage(mp)

	event := command.Handler(testRichMessage())

	assert.Equal(t, sendRichMessageFailedEventDef, event.EventDef)
	eventPayload := event.Payload.(SendRichMessageErrorOutput)
	assert.Equal(t, eventPayload.Error, "oh dear")
	assert.Equal(t, eventPayload.InputMessage, testRichMessageStruct())
}

func TestSendRichMessageHandlerShouldLogErrorWhenRichMessageSenderReturnsError(t *testing.T) {
	loggertest.Init("DEBUG")
	loggertest.ClearLogMessages()

	mp := mockRichMessageSender{
		sendRichMessage: func(rm client.RichMessage) (string, string, error) {
			return "","", errors.New("oh dear")
		},
	}

	cmd := SendRichMessage(mp)
	cmd.Handler(testRichMessage())

	msgs := loggertest.GetLogMessages()

	require.Equal(t, 1, len(msgs))
	assert.Equal(t, loggertest.LogLevelError, msgs[0].Level)
	assert.Equal(t, "error sending rich message: oh dear", msgs[0].Message)

	loggertest.Reset()
}

func TestPostMessageCallsMessagePoster(t *testing.T) {
	var sentMessage client.RichMessage
	mp := mockRichMessageSender{
		sendRichMessage: func(rm client.RichMessage) (string, string, error) {
			sentMessage = rm
			return "", "",nil
		},
	}

	command := SendRichMessage(mp)

	command.Handler(testRichMessage())

	// check a few fields to reassure ourselves that what got sent through was what was expected ... testing all would be testing json.Unmarshal:
	assert.Equal(t, "channel id", sentMessage.ChannelID)
	assert.Equal(t, "field value", sentMessage.Attachments[0].Fields[0].Value)
	assert.Equal(t, "action url", sentMessage.Attachments[0].Actions[0].URL)
}

func TestPostMessageReturnsEventWithInputMessageAsPayload(t *testing.T) {
	mp := mockRichMessageSender{
		sendRichMessage: func(rm client.RichMessage) (string, string, error) {
			return "AB45787HU", "1234.5678",nil
		},
	}
	command := SendRichMessage(mp)

	event := command.Handler(testRichMessage())
	im := event.Payload.(map[string]string)

	assert.Equal(t, richMessageSentEventDef, event.EventDef)

	assert.Equal(t, "1234.5678", im["threadTimestamp"])
	assert.Equal(t, "AB45787HU", im["channelId"])
}

func TestWiring(t *testing.T) {
	slack := NewMockSlack()
	slack.SendRichMessageFunc = func(rm client.RichMessage) (string, string, error) {
		return "", "", nil
	}

	command := SendRichMessage(slack)

	event := command.Handler(testRichMessage())

	assert.Equal(t, richMessageSentEventDef, event.EventDef)
}

func testRichMessage() []byte {
	return []byte(`{
		"channel":   "channel id",
		"thread_ts": "timestamp",
		"attachments": [
			{
				"fallback": "fallback",
				"color":    "color",
				"title":    "title",
				"text":     "text",
				"fields": [
					{
						"title": "field title",
						"value": "field value",
						"short": true
					}
				],
				"actions": [
					{
						"type":  "action type",
						"text":  "action text",
						"url":   "action url",
						"style": "action style"
					}
				]
			}
	    ]
	}`)
}

func testRichMessageStruct() client.RichMessage {
	var rm client.RichMessage
	json.Unmarshal(testRichMessage(), &rm)
	return rm
}

type mockRichMessageSender struct {
	sendRichMessage func(rm client.RichMessage) (string, string, error)
}

func (m mockRichMessageSender) SendRichMessage(rm client.RichMessage) (string, string, error) {
	if m.sendRichMessage != nil {
		return m.sendRichMessage(rm)
	}
	return "", "", nil
}
