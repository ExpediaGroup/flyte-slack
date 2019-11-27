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

var BroadcastMockSlack *MockSlack

func BeforeBroadcast() {

	BroadcastMockSlack = NewMockSlack()
	loggertest.Init("DEBUG")
}

func AfterBroadcast() {
	loggertest.Reset()
}

func TestBroadcastCommandIsPopulated(t *testing.T) {

	BeforeBroadcast()
	defer AfterBroadcast()

	command := Broadcast(BroadcastMockSlack)

	assert.Equal(t, "Broadcast", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "BroadcastSent", command.OutputEvents[0].Name)
	assert.Equal(t, "BroadcastFailed", command.OutputEvents[1].Name)
}

func TestBroadcastHandlerCallsSlackBroadcast(t *testing.T) {

	BeforeBroadcast()
	defer AfterBroadcast()

	handler := Broadcast(BroadcastMockSlack).Handler

	handler([]byte(`{"message": "yo"}`))

	calls := BroadcastMockSlack.BroadcastCalls
	require.Equal(t, 1, len(calls))
	assert.Equal(t, "yo", calls[0])
}

func TestBroadcastReturnsBroadcastSentEvent(t *testing.T) {

	BeforeBroadcast()
	defer AfterBroadcast()

	handler := Broadcast(BroadcastMockSlack).Handler

	event := handler([]byte(`{"message": "yo"}`))

	output := event.Payload.(BroadcastOutput)
	assert.Equal(t, "BroadcastSent", event.EventDef.Name)
	assert.Equal(t, "yo", output.Message)
}

func TestBroadcastHandleInvalidJsonInput(t *testing.T) {

	BeforeBroadcast()
	defer AfterBroadcast()

	handler := Broadcast(BroadcastMockSlack).Handler

	event := handler([]byte(`--- invalid json input ---`))

	output, ok := event.Payload.(string)
	require.True(t, ok)
	assert.Contains(t, output, "input is not valid: invalid character")
}

func TestBroadcastHandleInputWithMissingMessage(t *testing.T) {

	BeforeBroadcast()
	defer AfterBroadcast()

	handler := Broadcast(BroadcastMockSlack).Handler

	event := handler([]byte(`{}`))

	output := event.Payload.(BroadcastErrorOutput)
	assert.Equal(t, "BroadcastFailed", event.EventDef.Name)
	assert.Equal(t, "missing message field", output.Error)
}
