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

package client

import (
	"errors"
	"fmt"
	"github.com/HotelsDotCom/go-logger/loggertest"
	s "github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

var SlackImpl Slack
var SlackMockClient *MockClient

func Before(t *testing.T) {

	loggertest.Init("DEBUG")
	SlackImpl = NewSlack("token")
	SlackMockClient = NewMockClient(t)
	SlackImpl.(*slack).client = SlackMockClient
}

func After() {
	loggertest.Reset()
}

func TestSendMessage(t *testing.T) {

	Before(t)
	defer After()

	SlackImpl.SendMessage("the message", "channel id", "now")

	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
	assert.Equal(t, "the message", SlackMockClient.OutgoingMessages["channel id"][0].Text)

}

func TestSendRichMessage(t *testing.T) {
	Before(t)
	defer After()

	var pmp s.PostMessageParameters
	var ch, txt string
	SlackMockClient.PostMessageFunc = func(channel, text string, params s.PostMessageParameters) (string, string, error) {
		ch = channel
		txt = text
		pmp = params
		return "", "", nil
	}

	rm := RichMessage{
		Parse:           "pass?",
		ThreadTimestamp: "now",
		ReplyBroadcast:  true,
		LinkNames:       3,
		Attachments:     []s.Attachment{},
		UnfurlLinks:     true,
		UnfurlMedia:     true,
		IconURL:         "somewhere",
		IconEmoji:       "over the rainbow",
		Markdown:        true,
		EscapeText:      true,
		ChannelID:       "channel id",
		Text:            "hello?",
	}

	SlackImpl.SendRichMessage(rm)

	// args
	assert.Equal(t, "channel id", ch)
	assert.Equal(t, "hello?", txt)

	// PostMessageParameters values
	assert.True(t, pmp.AsUser)
	assert.Equal(t, "pass?", pmp.Parse)
	assert.Equal(t, "now", pmp.ThreadTimestamp)
	assert.True(t, pmp.ReplyBroadcast)
	assert.Equal(t, 3, pmp.LinkNames)
	assert.Equal(t, []s.Attachment{}, pmp.Attachments)
	assert.True(t, pmp.UnfurlLinks)
	assert.True(t, pmp.UnfurlMedia)
	assert.Equal(t, "somewhere", pmp.IconURL)
	assert.Equal(t, "over the rainbow", pmp.IconEmoji)
	assert.True(t, pmp.Markdown)
	assert.True(t, pmp.EscapeText)
}

func TestSendRichMessageShouldLogOnSuccess(t *testing.T) {
	Before(t)
	defer After()

	loggertest.ClearLogMessages()

	SlackMockClient.PostMessageFunc = func(channel, text string, params s.PostMessageParameters) (string, string, error) {
		return "", "", nil
	}

	rm := RichMessage{ChannelID: "channel id", Text: "hello?", ThreadTimestamp: "now"}
	SlackImpl.SendRichMessage(rm)

	msgs := loggertest.GetLogMessages()
	require.Equal(t, 1, len(msgs))
	assert.Equal(t, loggertest.LogLevelInfo, msgs[0].Level)
	assert.True(t, strings.HasPrefix(msgs[0].Message, "rich message="), "expected message to start with \"rich message=\", actual: %q", msgs[0].Message)
}

func TestSendRichMessageShouldReturnErrorOnFailure(t *testing.T) {
	Before(t)
	defer After()

	SlackMockClient.PostMessageFunc = func(channel, text string, params s.PostMessageParameters) (string, string, error) {
		return "", "", errors.New("barf")
	}

	rm := RichMessage{ChannelID: "channel id", Text: "hello?", ThreadTimestamp: "now"}
	err := SlackImpl.SendRichMessage(rm)

	require.NotNil(t, err)
	errMsg := err.Error()
	assert.True(t, strings.HasPrefix(errMsg, "cannot send rich message="), "expected error to start with \"cannot send rich message:\", actual: %q", errMsg)
	assert.True(t, strings.HasSuffix(errMsg, ": barf"), "expected message to end with \": barf\", actual: %q", errMsg)
}

func TestIncomingMessages(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	u := &s.User{
		ID:   "user-id-123",
		Name: "kfoox",
		Profile: s.UserProfile{
			Title:     "boss",
			Email:     "k@example.com",
			FirstName: "Karl",
			LastName:  "Foox",
		},
	}
	SlackMockClient.AddMockGetUserInfoCall("user-id-123", u, nil)

	incomingMessages := SlackImpl.IncomingMessages()
	sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id-123", "now", "thread")
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-incomingMessages:
		assert.Equal(t, "ReceivedMessage", msg.EventDef.Name)
		payload := msg.Payload.(messageEvent)
		assert.Equal(t, "id-abc", payload.ChannelId)
		assert.Equal(t, "hello there ...", payload.Message)
		assert.Equal(t, "user-id-123", payload.User.Id)
		assert.Equal(t, "kfoox", payload.User.Name)
		assert.Equal(t, "boss", payload.User.Title)
		assert.Equal(t, "k@example.com", payload.User.Email)
		assert.Equal(t, "Karl", payload.User.FirstName)
		assert.Equal(t, "Foox", payload.User.LastName)
		assert.Equal(t, "now", payload.Timestamp)
		assert.Equal(t, "thread", payload.ThreadTimestamp)
	default:
		assert.Fail(t, "expected message event")
	}
	// Test whether Thread is properly populated or not
	tests := []struct {
		name                    string
		inputTimestamp          string
		inputThreadTimestamp    string
		expectedThreadTimestamp string
	}{
		{
			name:                    "both Timestamp and ThreadTimestamp are defined",
			inputTimestamp:          "ts",
			inputThreadTimestamp:    "tts",
			expectedThreadTimestamp: "tts",
		},
		{
			name:                    "ThreadTimestamp is defined only",
			inputTimestamp:          "ts",
			inputThreadTimestamp:    "",
			expectedThreadTimestamp: "ts",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			SlackMockClient.AddMockGetUserInfoCall("user-id-123", u, nil)
			sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id-123", test.inputTimestamp, test.inputThreadTimestamp)
			time.Sleep(50 * time.Millisecond)

			select {
			case msg := <-incomingMessages:
				assert.Equal(t, "ReceivedMessage", msg.EventDef.Name)
				payload := msg.Payload.(messageEvent)
				assert.Equal(t, test.expectedThreadTimestamp, payload.ThreadTimestamp)
			default:
				assert.Fail(t, "expected message event")
			}
		})
	}
}

func TestIncomingMessagesLogging(t *testing.T) {

	// given
	Before(t)
	defer After()
	loggertest.ClearLogMessages()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	u := &s.User{
		ID:   "user-id-123",
		Name: "kfoox",
		Profile: s.UserProfile{
			Title:     "boss",
			Email:     "k@example.com",
			FirstName: "Karl",
			LastName:  "Foox",
		},
	}
	SlackMockClient.AddMockGetUserInfoCall("user-id-123", u, nil)

	// when
	sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id-123", "", "")
	time.Sleep(50 * time.Millisecond)

	// then
	msgs := loggertest.GetLogMessages()
	require.Equal(t, 1, len(msgs))

	assert.Equal(t, "received message=hello there ... in channel=id-abc", msgs[0].Message)
	assert.Equal(t, loggertest.LogLevelDebug, msgs[0].Level)
}

// --- helpers ---

// this simulates messages coming from slack
func sendSlackMessage(slackImpl Slack, text, channel, userId, timestamp, threadTimestamp string) {

	data := &s.MessageEvent{}
	data.Text = text
	data.Channel = channel
	data.User = userId
	data.Timestamp = timestamp
	data.ThreadTimestamp = threadTimestamp
	messageEvent := s.RTMEvent{Type: "message", Data: data}

	slackImpl.(*slack).incomingEvents <- messageEvent
}

// --- mock client ---

type MockClient struct {
	t *testing.T
	// Slice of mocked get user info functions (call to GetUserInfo will pop from slice)
	GetUserInfoFns []func(userId string) (*s.User, error)
	// map stores all the sent messages by channelId (key is channelId)
	OutgoingMessages map[string][]*s.OutgoingMessage
	// Slice of rich messages
	PostMessageFunc func(channel, text string, params s.PostMessageParameters) (string, string, error)
}

func NewMockClient(t *testing.T) *MockClient {

	m := &MockClient{t: t}
	m.GetUserInfoFns = []func(userId string) (*s.User, error){}
	m.OutgoingMessages = make(map[string][]*s.OutgoingMessage)
	m.PostMessageFunc = func(channel, text string, params s.PostMessageParameters) (string, string, error) {
		return "", "", nil
	}
	return m
}

// --- config methods ---

func (m *MockClient) AddMockGetUserInfoCall(expectedUserId string, returnUser *s.User, returnError error) {

	x := func(userId string) (*s.User, error) {
		errMessage := fmt.Sprintf("GetUserInfo call expected %s userId got %s", expectedUserId, userId)
		require.Equal(m.t, expectedUserId, userId, errMessage)
		return returnUser, returnError
	}
	m.GetUserInfoFns = append(m.GetUserInfoFns, x)
}

// --- implementation ---

func (m *MockClient) GetUserInfo(userId string) (*s.User, error) {

	require.NotEmpty(m.t, m.GetUserInfoFns, "Unexpected (not mocked) call to get user info")

	var fn func(string) (*s.User, error)
	fn, m.GetUserInfoFns = m.GetUserInfoFns[0], m.GetUserInfoFns[1:]
	return fn(userId)
}

func (m *MockClient) NewOutgoingMessage(message, channelId string, options ...s.RTMsgOption) *s.OutgoingMessage {

	return &s.OutgoingMessage{
		ID:      666,
		Type:    "message",
		Channel: channelId,
		Text:    message,
	}
}

func (m *MockClient) SendMessage(message *s.OutgoingMessage) {
	m.OutgoingMessages[message.Channel] = append(m.OutgoingMessages[message.Channel], message)
}

func (m *MockClient) PostMessage(channel, text string, params s.PostMessageParameters) (string, string, error) {
	return m.PostMessageFunc(channel, text, params)
}

// --- backup helper ---

type inMemoryBackup struct {
	channelIds []string
}

func (i *inMemoryBackup) Backup(channelIds []string) {
	i.channelIds = channelIds
}

func (i *inMemoryBackup) Load() []string {
	return i.channelIds
}
