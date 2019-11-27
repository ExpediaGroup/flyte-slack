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

var JoinMockSlack *MockSlack

func BeforeJoin() {

	JoinMockSlack = NewMockSlack()
	loggertest.Init("DEBUG")
}

func AfterJoin() {
	loggertest.Reset()
}

func TestJoinChannelCommandIsPopulated(t *testing.T) {

	command := JoinChannel(JoinMockSlack)

	assert.Equal(t, "JoinChannel", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "ChannelJoined", command.OutputEvents[0].Name)
	assert.Equal(t, "JoinChannelFailed", command.OutputEvents[1].Name)
}

func TestJoinChannelReturnsChannelJoinedEvent(t *testing.T) {

	BeforeJoin()
	defer AfterJoin()

	handler := JoinChannel(JoinMockSlack).Handler
	event := handler([]byte(`{"channelId": "CB8765XY"}`))

	output := event.Payload.(JoinChannelOutput)
	assert.Equal(t, "ChannelJoined", event.EventDef.Name)
	assert.Equal(t, "CB8765XY", output.ChannelId)
}

func TestJoinChannelCallsJoinSlackMethod(t *testing.T) {

	BeforeJoin()
	defer AfterJoin()

	handler := JoinChannel(JoinMockSlack).Handler
	handler([]byte(`{"channelId": "CB8765XY"}`))

	calls := JoinMockSlack.JoinChannelCalls
	require.Equal(t, 1, len(calls))
	assert.Equal(t, "CB8765XY", calls[0])
}

func TestJoinChannelHandleInvalidJsonInput(t *testing.T) {

	BeforeJoin()
	defer AfterJoin()

	handler := JoinChannel(JoinMockSlack).Handler
	event := handler([]byte(`--- invalid json input ---`))

	output, ok := event.Payload.(string)
	require.True(t, ok)
	assert.Contains(t, output, "input is not valid: invalid character")
}

func TestJoinChannelHandleInputWithMissingChannelId(t *testing.T) {

	BeforeJoin()
	defer AfterJoin()

	handler := JoinChannel(JoinMockSlack).Handler
	event := handler([]byte(`{}`))

	output := event.Payload.(JoinChannelFailedOutput)
	assert.Equal(t, "JoinChannelFailed", event.EventDef.Name)
	assert.Equal(t, "missing channel id field", output.Error)
}
