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

var LeaveMockSlack *MockSlack

func BeforeLeave() {

	LeaveMockSlack = NewMockSlack()
	loggertest.Init("DEBUG")
}

func AfterLeave() {
	loggertest.Reset()
}

func TestLeaveChannelCommandIsPopulated(t *testing.T) {

	command := LeaveChannel(LeaveMockSlack)

	assert.Equal(t, "LeaveChannel", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "ChannelLeft", command.OutputEvents[0].Name)
	assert.Equal(t, "LeaveChannelFailed", command.OutputEvents[1].Name)
}

func TestLeaveChannelReturnsChannelLeftEvent(t *testing.T) {

	BeforeLeave()
	defer AfterLeave()

	handler := LeaveChannel(LeaveMockSlack).Handler

	event := handler([]byte(`{"channelId": "AB8765XY"}`))

	output := event.Payload.(LeaveChannelOutput)
	assert.Equal(t, "ChannelLeft", event.EventDef.Name)
	assert.Equal(t, "AB8765XY", output.ChannelId)
}

func TestLeaveChannelCallsLeaveSlackMethod(t *testing.T) {

	BeforeLeave()
	defer AfterLeave()

	handler := LeaveChannel(LeaveMockSlack).Handler

	handler([]byte(`{"channelId": "AB8765XY"}`))

	calls := LeaveMockSlack.LeaveChannelCalls
	require.Equal(t, 1, len(calls))
	assert.Equal(t, "AB8765XY", calls[0])
}

func TestLeaveChannelHandleInvalidJsonInput(t *testing.T) {

	BeforeLeave()
	defer AfterLeave()

	handler := LeaveChannel(LeaveMockSlack).Handler

	event := handler([]byte(`--- invalid json input ---`))

	output, ok := event.Payload.(string)
	require.True(t, ok)
	assert.Contains(t, output, "input is not valid: invalid character")
}

func TestLeaveChannelHandleInputWithMissingChannelId(t *testing.T) {

	BeforeLeave()
	defer AfterLeave()

	handler := LeaveChannel(LeaveMockSlack).Handler

	event := handler([]byte(`{}`))

	output := event.Payload.(LeaveChannelFailedOutput)
	assert.Equal(t, "LeaveChannelFailed", event.EventDef.Name)
	assert.Equal(t, "missing channel id field", output.Error)
}
