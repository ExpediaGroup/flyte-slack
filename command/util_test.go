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
	"github.com/HotelsDotCom/flyte-slack/client"
)

type MockSlack struct {
	SendMessageCalls    map[string][]string
	BroadcastCalls      []string
	JoinChannelCalls    []string
	LeaveChannelCalls   []string
	JoinedChannelsCalls int
	SendRichMessageFunc func(rm client.RichMessage) error
}

func NewMockSlack() *MockSlack {

	m := &MockSlack{}
	m.SendMessageCalls = make(map[string][]string)
	return m
}

func (m *MockSlack) SendMessage(message, channelId, _ string) {
	m.SendMessageCalls[channelId] = append(m.SendMessageCalls[channelId], message)
}

func (m *MockSlack) SendRichMessage(rm client.RichMessage) error {
	return m.SendRichMessageFunc(rm)
}

func (m *MockSlack) Broadcast(message string) {
	m.BroadcastCalls = append(m.BroadcastCalls, message)
}

func (m *MockSlack) JoinChannel(channelId string) {
	m.JoinChannelCalls = append(m.JoinChannelCalls, channelId)
}

func (m *MockSlack) LeaveChannel(channelId string) {
	m.LeaveChannelCalls = append(m.LeaveChannelCalls, channelId)
}

func (m *MockSlack) JoinedChannels() []string {

	m.JoinedChannelsCalls++
	return m.JoinChannelCalls
}

func (m *MockSlack) IncomingMessages() <-chan flyte.Event {
	return make(chan flyte.Event)
}
