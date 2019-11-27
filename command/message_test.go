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
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var MessageMockSlack *MockSlack

func BeforeMessage() {

	MessageMockSlack = NewMockSlack()
	loggertest.Init("DEBUG")
}

func AfterMessage() {
	loggertest.Reset()
}

func TestSendMessageCommandIsPopulated(t *testing.T) {

	command := SendMessage(MessageMockSlack)

	assert.Equal(t, "SendMessage", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "MessageSent", command.OutputEvents[0].Name)
	assert.Equal(t, "SendMessageFailed", command.OutputEvents[1].Name)
}

func TestSendsMessageSendsMessageToSlack(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := SendMessage(MessageMockSlack).Handler
	handler([]byte(`{"message": "hello from flyte", "channelId": "abc-channel"}`))

	calls := MessageMockSlack.SendMessageCalls
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, len(calls["abc-channel"]))
	assert.Equal(t, "hello from flyte", calls["abc-channel"][0])
}

func TestSendMessageReturnsMessageSentEvent(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := SendMessage(MessageMockSlack).Handler
	event := handler([]byte(`{"message": "yo", "channelId": "xyz"}`))

	output := event.Payload.(SendMessageOutput)
	assert.Equal(t, "MessageSent", event.EventDef.Name)
	assert.Equal(t, "yo", output.Message)
	assert.Equal(t, "xyz", output.ChannelId)
}

func TestSendMessageHandleInvalidJsonInput(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := SendMessage(MessageMockSlack).Handler
	event := handler([]byte(`--- invalid json input ---`))

	output, ok := event.Payload.(string)
	require.True(t, ok)
	assert.Contains(t, output, "input is not valid: invalid character")
}

func TestSendMessageHandleInputWithMissingMessage(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := SendMessage(MessageMockSlack).Handler
	event := handler([]byte(`{"channelId": "UXB456Y"}`))

	output := event.Payload.(SendMessageErrorOutput)
	assert.Equal(t, "SendMessageFailed", event.EventDef.Name)
	assert.Equal(t, "missing message field", output.Error)
}

func TestSendMessageHandleInputWithMissingChannelId(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := SendMessage(MessageMockSlack).Handler
	event := handler([]byte(`{"message": "oh, the channel id is missing"}`))

	output := event.Payload.(SendMessageErrorOutput)
	assert.Equal(t, "SendMessageFailed", event.EventDef.Name)
	assert.Equal(t, "missing channel id field", output.Error)
}

func TestSendMessageHandleInputWithMissingMessageAndChannelId(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := SendMessage(MessageMockSlack).Handler
	event := handler([]byte(`{}`))

	output := event.Payload.(SendMessageErrorOutput)
	assert.Equal(t, "SendMessageFailed", event.EventDef.Name)
	assert.Equal(t, "missing message field, missing channel id field", output.Error)
}
