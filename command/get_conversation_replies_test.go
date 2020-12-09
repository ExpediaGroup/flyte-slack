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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConversationRepliesCommandIsPopulated(t *testing.T) {
	command := GetConversationReplies(MessageMockSlack)

	assert.Equal(t, "GetConversationReplies", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "GetConversationReplies", command.OutputEvents[0].Name)
	assert.Equal(t, "GetConversationRepliesFailed", command.OutputEvents[1].Name)
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithInvalidJSON(t *testing.T) {
	cmd := GetConversationReplies(nil)

	event := cmd.Handler([]byte(`.`))

	fatalEventDef := flyte.NewFatalEvent("").EventDef
	assert.Equal(t, fatalEventDef, event.EventDef)
	assert.Contains(t, event.Payload.(string), "invalid input: ")
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithMissingTimestamp(t *testing.T) {
	BeforeMessage()
	defer AfterMessage()

	handler := GetConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"channelId": "UXB456Y"}`))

	output := event.Payload.(GetConversationRepliesErrorOutput)
	assert.Equal(t, "GetConversationRepliesFailed", event.EventDef.Name)
	assert.Equal(t, "missing threadTimestamp field", output.Error)
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithMissingChannelId(t *testing.T) {
	BeforeMessage()
	defer AfterMessage()

	handler := GetConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"threadTimestamp": "333333"}`))

	output := event.Payload.(GetConversationRepliesErrorOutput)
	assert.Equal(t, "GetConversationRepliesFailed", event.EventDef.Name)
	assert.Equal(t, "missing channel id field", output.Error)
}

func TestGetConversationRepliesShouldReturnFatalErrorEventWhenCalledWithMissingChannelIdAndTimestamp(t *testing.T) {
	BeforeMessage()
	defer AfterMessage()

	handler := GetConversationReplies(MessageMockSlack).Handler
	event := handler([]byte(`{"abc": "123"}`))

	output := event.Payload.(GetConversationRepliesErrorOutput)
	assert.Equal(t, "GetConversationRepliesFailed", event.EventDef.Name)
	assert.Equal(t, "missing threadTimestamp field, missing channel id field", output.Error)
}
