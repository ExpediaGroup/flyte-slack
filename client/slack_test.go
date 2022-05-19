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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ExpediaGroup/flyte-client/flyte"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

var SlackImpl Slack
var SlackMockClient *MockClient

func Before(t *testing.T) {
	SlackImpl = NewSlack("token")
	SlackMockClient = NewMockClient(t)
	SlackImpl.(*slackClient).client = SlackMockClient
}

func TestSendMessage(t *testing.T) {
	Before(t)

	SlackImpl.SendMessage("the message", "channel id", "now")

	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
	assert.Equal(t, "the message", SlackMockClient.OutgoingMessages["channel id"][0].Text)

}

func TestSendRichMessage(t *testing.T) {
	Before(t)

	var ch string
	SlackMockClient.PostMessageFunc = func(channel string, opts ...slack.MsgOption) (string, string, error) {
		ch = channel
		return "", "", nil
	}

	rm := RichMessage{
		Username:        "my_bot",
		Parse:           "pass?",
		ThreadTimestamp: "now",
		ReplyBroadcast:  true,
		LinkNames:       3,
		Attachments:     []slack.Attachment{},
		UnfurlLinks:     true,
		UnfurlMedia:     true,
		IconURL:         "somewhere",
		IconEmoji:       "over the rainbow",
		Markdown:        true,
		EscapeText:      true,
		ChannelID:       "channel id",
		Text:            "hello?",
	}

	_, _, err := SlackImpl.SendRichMessage(rm)
	require.NoError(t, err)

	// args
	assert.Equal(t, "channel id", ch)
}

func TestSendRichMessageShouldReturnErrorOnFailure(t *testing.T) {
	Before(t)

	SlackMockClient.PostMessageFunc = func(channel string, opts ...slack.MsgOption) (string, string, error) {
		return "", "", errors.New("barf")
	}

	rm := RichMessage{ChannelID: "channel id", Text: "hello?", ThreadTimestamp: "now"}
	_, _, err := SlackImpl.SendRichMessage(rm)

	require.NotNil(t, err)
	errMsg := err.Error()
	assert.True(t, strings.HasPrefix(errMsg, "cannot send rich message="), "expected error to start with \"cannot send rich message:\", actual: %q", errMsg)
	assert.True(t, strings.HasSuffix(errMsg, ": barf"), "expected message to end with \": barf\", actual: %q", errMsg)
}

func TestIncomingMessages(t *testing.T) {

	Before(t)
	c1 := &slack.Channel{}
	c1.Name = "name-abc"
	u := &slack.User{
		ID:   "user-id-123",
		Name: "kfoox",
		Profile: slack.UserProfile{
			Title:     "boss",
			Email:     "k@example.com",
			FirstName: "Karl",
			LastName:  "Foox",
		},
	}
	slackReplies := []slack.Reply{
		{
			Timestamp: "123",
			User:      "Karl",
		},
	}

	SlackMockClient.AddMockGetUserInfoCall("user-id-123", u, nil)
	incomingMessages := SlackImpl.IncomingMessages()
	sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id-123", "now", "thread", 2, slackReplies)
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
		assert.Equal(t, 2, payload.ReplyCount)
		assert.Equal(t, "123", payload.Replies[0].Timestamp)
		assert.Equal(t, "Karl", payload.Replies[0].User)

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
			sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id-123", test.inputTimestamp, test.inputThreadTimestamp, 2, slackReplies)
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

func TestReactionEvents(t *testing.T) {
	// Given
	Before(t)
	u := &slack.User{
		ID:   "u-foo",
		Name: "kfoox",
		Profile: slack.UserProfile{
			Title:     "boss",
			Email:     "k@example.com",
			FirstName: "Karl",
			LastName:  "Foox",
		},
	}
	SlackMockClient.AddMockGetUserInfoCall("u-foo", u, nil)
	SlackMockClient.AddMockGetUserInfoCall("u-foo", u, nil)

	// When
	reaction, err := newReactionAddedEvent("u-foo", "my-item-type",
		"id-abc", "123.1", "create-a-ticket", "123.2")
	require.NoError(t, err)
	SlackImpl.(*slackClient).incomingEvents <- slack.RTMEvent{Type: "reaction_added", Data: &reaction}

	// Then
	incomingMessages := SlackImpl.IncomingMessages()
	select {
		case msg := <-incomingMessages:
			u := user{
				Id:        "u-foo",
				Name:      "kfoox",
				Email:     "k@example.com",
				Title:     "boss",
				FirstName: "Karl",
				LastName:  "Foox",
			}
			want := flyte.Event{
				EventDef: flyte.EventDef{Name: "ReactionAdded"},
				Payload: reactionEvent{
					Type:     "reaction_added",
					User:     u,
					ItemUser: u,
					Item: reactionItem{
						Type:      "my-item-type",
						Channel:   "id-abc",
						Timestamp: "123.1",
					},
					Reaction:       "create-a-ticket",
					EventTimestamp: "123.2",
				},
			}
			assert.Equal(t, want, msg)

		case <-time.After(250 * time.Millisecond):
			assert.Fail(t, "Timed out while waiting for an expected reaction's event!")
	}
}

// --- helpers ---

// this simulates messages coming from slack
func sendSlackMessage(slackImpl Slack, text, channel, userId, timestamp, threadTimestamp string, replyCount int, replies []slack.Reply) {

	data := &slack.MessageEvent{
		Msg: slack.Msg{
			Text:            text,
			Channel:         channel,
			User:            userId,
			Timestamp:       timestamp,
			ThreadTimestamp: threadTimestamp,
			ReplyCount:      replyCount,
			Replies:         replies,
		},
	}

	messageEvent := slack.RTMEvent{Type: "message", Data: data}

	slackImpl.(*slackClient).incomingEvents <- messageEvent
}

func newReactionAddedEvent(userId, itemType, channel, itemTs, reaction, ts string) (slack.ReactionAddedEvent, error) {
	data := fmt.Sprintf(`{
		"type": "reaction_added",
		"user": "%s",
		"item_user": "%s",
		"item": {
			"type": "%s",
			"channel": "%s",
			"ts": "%s"
		},
		"reaction": "%s",
		"event_ts": "%s"}`,
		userId, userId, itemType, channel, itemTs, reaction, ts)
	var r slack.ReactionAddedEvent
	err := json.Unmarshal([]byte(data), &r)
	return r, err
}

// --- mock client ---

type MockClient struct {
	t *testing.T
	// Slice of mocked get user info functions (call to GetUserInfo will pop from slice)
	GetUserInfoFns []func(userId string) (*slack.User, error)
	// map stores all the sent messages by channelId (key is channelId)
	OutgoingMessages map[string][]*slack.OutgoingMessage
	// Slice of rich messages
	PostMessageFunc func(channel string, opts ...slack.MsgOption) (string, string, error)
}

func NewMockClient(t *testing.T) *MockClient {

	m := &MockClient{t: t}
	m.GetUserInfoFns = []func(userId string) (*slack.User, error){}
	m.OutgoingMessages = make(map[string][]*slack.OutgoingMessage)
	m.PostMessageFunc = func(channel string, params ...slack.MsgOption) (string, string, error) {
		return "", "", nil
	}

	return m
}

// --- config methods ---

func (m *MockClient) AddMockGetUserInfoCall(expectedUserId string, returnUser *slack.User, returnError error) {

	x := func(userId string) (*slack.User, error) {
		errMessage := fmt.Sprintf("GetUserInfo call expected %s userId got %s", expectedUserId, userId)
		require.Equal(m.t, expectedUserId, userId, errMessage)
		return returnUser, returnError
	}
	m.GetUserInfoFns = append(m.GetUserInfoFns, x)
}

// --- implementation ---

func (m *MockClient) GetUserInfo(userId string) (*slack.User, error) {

	require.NotEmpty(m.t, m.GetUserInfoFns, "Unexpected (not mocked) call to get user info")

	var fn func(string) (*slack.User, error)
	fn, m.GetUserInfoFns = m.GetUserInfoFns[0], m.GetUserInfoFns[1:]
	return fn(userId)
}

func (m *MockClient) NewOutgoingMessage(message, channelId string, options ...slack.RTMsgOption) *slack.OutgoingMessage {

	return &slack.OutgoingMessage{
		ID:      666,
		Type:    "message",
		Channel: channelId,
		Text:    message,
	}
}

func (m *MockClient) SendMessage(message *slack.OutgoingMessage) {
	m.OutgoingMessages[message.Channel] = append(m.OutgoingMessages[message.Channel], message)
}

func (m *MockClient) PostMessage(channel string, opts ...slack.MsgOption) (string, string, error) {
	return m.PostMessageFunc(channel, opts...)
}

func (m *MockClient) GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error) {
	return nil, "", err
}
