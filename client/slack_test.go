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
	s "github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"testing"
	"time"
	"strings"
)

var SlackImpl Slack
var SlackBackup Backup
var SlackMockClient *MockClient

func Before(t *testing.T) {

	loggertest.Init("DEBUG")
	SlackBackup = NewInMemoryBackup()
	SlackImpl = NewSlack(SlackBackup, "token", "")
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

	rm := RichMessage{ChannelID: "channel id", Text:"hello?", ThreadTimestamp:"now"}
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

	rm := RichMessage{ChannelID: "channel id", Text:"hello?", ThreadTimestamp:"now"}
	err := SlackImpl.SendRichMessage(rm)

	require.NotNil(t, err)
	errMsg := err.Error()
	assert.True(t, strings.HasPrefix(errMsg, "cannot send rich message="), "expected error to start with \"cannot send rich message:\", actual: %q", errMsg)
	assert.True(t, strings.HasSuffix(errMsg, ": barf"), "expected message to end with \": barf\", actual: %q", errMsg)
}

func TestBroadcastSendsMessageToAllJoinedChannels(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-1"
	SlackMockClient.AddMockGetChannelInfoCall("id-1", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-1", nil)
	c2 := &s.Channel{}
	c2.Name = "name-2"
	SlackMockClient.AddMockGetChannelInfoCall("id-2", c2, nil)
	SlackMockClient.AddMockJoinChannelCall("name-2", nil)
	c3 := &s.Channel{}
	c3.Name = "name-3"
	SlackMockClient.AddMockGetChannelInfoCall("id-3", c3, nil)
	SlackMockClient.AddMockJoinChannelCall("name-3", nil)

	SlackImpl.JoinChannel("id-1")
	SlackImpl.JoinChannel("id-2")
	SlackImpl.JoinChannel("id-3")

	SlackImpl.Broadcast("sending message to all joined channels")

	// 3 rooms, 2 messages in each room (first one is join message)
	require.Equal(t, 3, len(SlackMockClient.OutgoingMessages))
	require.Equal(t, 2, len(SlackMockClient.OutgoingMessages["id-1"]))
	require.Equal(t, 2, len(SlackMockClient.OutgoingMessages["id-2"]))
	require.Equal(t, 2, len(SlackMockClient.OutgoingMessages["id-3"]))

	assert.Equal(t, "sending message to all joined channels", SlackMockClient.OutgoingMessages["id-1"][1].Text)
	assert.Equal(t, "sending message to all joined channels", SlackMockClient.OutgoingMessages["id-2"][1].Text)
	assert.Equal(t, "sending message to all joined channels", SlackMockClient.OutgoingMessages["id-3"][1].Text)
}

func TestJoinPersistsChannelAndSendsJoinMessage(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)

	SlackImpl.JoinChannel("id-abc")

	assert.Contains(t, SlackImpl.JoinedChannels(), "id-abc")
	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
	assert.Equal(t, "Hello! I've joined this room ...", SlackMockClient.OutgoingMessages["id-abc"][0].Text)
}

func TestJoinBacksUpChannels(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)

	SlackImpl.JoinChannel("id-abc")

	backedUpChannels := SlackBackup.Load()
	require.Equal(t, 1, len(backedUpChannels))
	assert.Equal(t, "id-abc", backedUpChannels[0])
}

func TestJoinDoesNotBacksUpChannelsIfJoinCallErrors(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", errors.New("test join error"))

	SlackImpl.JoinChannel("id-abc")

	backedUpChannels := SlackBackup.Load()
	require.Equal(t, 0, len(backedUpChannels))
}

func TestJoinDoesNotSendJoinMessageIfJoinCallErrors(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", errors.New("test join error"))

	SlackImpl.JoinChannel("id-abc")

	require.Equal(t, 0, len(SlackMockClient.OutgoingMessages))
}

func TestMultipleJoinCallsSendOnlyOneJoinMessage(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)

	SlackImpl.JoinChannel("id-abc")
	SlackImpl.JoinChannel("id-abc")
	SlackImpl.JoinChannel("id-abc")

	assert.Contains(t, SlackImpl.JoinedChannels(), "id-abc")
	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
	assert.Equal(t, "Hello! I've joined this room ...", SlackMockClient.OutgoingMessages["id-abc"][0].Text)
}

func TestLeaveRemovesChannelAndSendsLeaveMessage(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
	SlackMockClient.AddMockLeaveChannelCall("id-abc", false, nil)

	SlackImpl.JoinChannel("id-abc")
	SlackImpl.LeaveChannel("id-abc")

	assert.NotContains(t, SlackImpl.JoinedChannels(), "id-abc")
	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
	require.Equal(t, 2, len(SlackMockClient.OutgoingMessages["id-abc"]))
	assert.Equal(t, "Hello! I've joined this room ...", SlackMockClient.OutgoingMessages["id-abc"][0].Text)
	assert.Equal(t, "I'm leaving now, bye!", SlackMockClient.OutgoingMessages["id-abc"][1].Text)
}

func TestLeaveRemovesChannelFromBackup(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
	SlackMockClient.AddMockLeaveChannelCall("id-abc", false, nil)

	// join - channel is backed up
	SlackImpl.JoinChannel("id-abc")
	backedUpChannels := SlackBackup.Load()
	require.Equal(t, 1, len(backedUpChannels))

	// leave - channel is removed from back up
	SlackImpl.LeaveChannel("id-abc")
	backedUpChannels = SlackBackup.Load()
	require.Equal(t, 0, len(backedUpChannels))
}

// rtm.SendMessage (different than rtm.PostMessage - which is api http POST call)
// sends message only if the channel is joined. Therefore send message has to
// be called before leave channel (and it is safe as well - it won't send the message
// if the channel is not joined)
func TestLeaveSendsLeaveMessageBeforeItCallsLeaveChannel(t *testing.T) {

	Before(t)
	defer After()
	SlackMockClient.AddMockLeaveChannelCall("not-joined", true, nil)

	SlackImpl.LeaveChannel("not-joined")
	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
}

// same as test above, message has to be send before we leave channel
func TestLeaveSendsLeaveMessageIfLeaveCallErrors(t *testing.T) {

	Before(t)
	defer After()
	SlackMockClient.AddMockLeaveChannelCall("id-errors", false, errors.New("test leave error"))

	SlackImpl.LeaveChannel("id-errors")
	require.Equal(t, 1, len(SlackMockClient.OutgoingMessages))
}

func TestLeaveDoesNotRemoveChannelFromBackupIfLeaveCallErrors(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-errors-on-leave"
	SlackMockClient.AddMockGetChannelInfoCall("id-errors-on-leave", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-errors-on-leave", nil)
	SlackMockClient.AddMockLeaveChannelCall("id-errors-on-leave", false, errors.New("test leave error"))

	SlackImpl.JoinChannel("id-errors-on-leave")
	backedUpChannels := SlackBackup.Load()
	require.Equal(t, 1, len(backedUpChannels))

	SlackImpl.LeaveChannel("id-errors-on-leave")
	backedUpChannels = SlackBackup.Load()
	require.Equal(t, 1, len(backedUpChannels))
}

func TestIncomingMessages(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
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
	SlackImpl.JoinChannel("id-abc")
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
}

func TestIncomingMessagesLogging(t *testing.T) {

	// given
	Before(t)
	defer After()
	loggertest.ClearLogMessages()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
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
	SlackImpl.JoinChannel("id-abc")
	sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id-123", "", "")
	time.Sleep(50 * time.Millisecond)

	// then
	msgs := loggertest.GetLogMessages()
	require.Equal(t, 3, len(msgs))
	assert.Equal(t, "message=\"Hello! I've joined this room ...\" sent to channel=id-abc", msgs[0].Message)
	assert.Equal(t, loggertest.LogLevelInfo, msgs[0].Level)

	assert.Equal(t, "joined channel=id-abc", msgs[1].Message)
	assert.Equal(t, loggertest.LogLevelInfo, msgs[1].Level)

	assert.Equal(t, "received message=hello there ... in channel=id-abc", msgs[2].Message)
	assert.Equal(t, loggertest.LogLevelDebug, msgs[2].Level)
}

func TestIncomingMessagesOnChannelThatIsNotJoined(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)

	// join abc channel
	SlackImpl.JoinChannel("id-abc")
	incomingMessages := SlackImpl.IncomingMessages()

	// send message to xyz channel
	sendSlackMessage(SlackImpl, "hello there ...", "xyz", "user-id", "", "")
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-incomingMessages:
		assert.Fail(t, fmt.Sprintf("did not expect any message event, got %s: %v", msg.EventDef.Name, msg.Payload))
	default:
		// do not block
	}
}

func TestIncomingMessagesOnLeftChannel(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-abc"
	SlackMockClient.AddMockGetChannelInfoCall("id-abc", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-abc", nil)
	SlackMockClient.AddMockLeaveChannelCall("id-abc", false, nil)
	u := &s.User{ID:"user-id", Name: "kfoox"}
	SlackMockClient.AddMockGetUserInfoCall("user-id", u, nil)

	// join abc channel
	SlackImpl.JoinChannel("id-abc")
	incomingMessages := SlackImpl.IncomingMessages()

	sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id", "", "")
	time.Sleep(50 * time.Millisecond)

	// consume message
	select {
	case msg := <-incomingMessages:
		assert.Equal(t, "hello there ...", msg.Payload.(messageEvent).Message)
	default:
		assert.Fail(t, "expected message event")
	}

	// leave abc channel and send another message to that channel
	SlackImpl.LeaveChannel("id-abc")
	sendSlackMessage(SlackImpl, "hello there ...", "id-abc", "user-id", "", "")
	time.Sleep(50 * time.Millisecond)

	// no message should be received (we left the channel)
	select {
	case <-incomingMessages:
		assert.Fail(t, "did not expect any message event")
	default:
		// do not block
	}
}

func TestIncomingMessagesOnJoinedChannel(t *testing.T) {

	Before(t)
	defer After()
	c1 := &s.Channel{}
	c1.Name = "name-123"
	SlackMockClient.AddMockGetChannelInfoCall("id-123", c1, nil)
	SlackMockClient.AddMockJoinChannelCall("name-123", nil)
	u := &s.User{ID:"user-id", Name: "kfoox"}
	SlackMockClient.AddMockGetUserInfoCall("user-id", u, nil)

	// send message (no channel is joined yet)
	sendSlackMessage(SlackImpl, "hello there ...", "id-123", "user-id", "", "")

	// no message should be received (no channel is joined)
	incomingMessages := SlackImpl.IncomingMessages()
	time.Sleep(50 * time.Millisecond)

	select {
	case <-incomingMessages:
		assert.Fail(t, "did not expect any message event")
	default:
		// do not block
	}

	// join 123 channel and send message to that channel
	SlackImpl.JoinChannel("id-123")
	sendSlackMessage(SlackImpl, "one two three ...", "id-123", "user-id", "", "")
	time.Sleep(50 * time.Millisecond)

	// consume message
	select {
	case msg := <-incomingMessages:
		assert.Equal(t, "one two three ...", msg.Payload.(messageEvent).Message)
	default:
		//assert.Fail(t, "expected message event")
	}
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
	// Slice of mocked join channel functions (call to JoinChannel will pop from slice)
	JoinChannelFns []func(channelName string) error
	// Slice of mocked leave channel functions (call to LeaveChannel will pop from slice)
	LeaveChannelFns []func(channelId string) (notInChannel bool, err error)
	// Slice of mocked get channel info functions (call to GetChannelInfo will pop from slice)
	GetChannelInfoFns []func(channelId string) (*s.Channel, error)
	// Slice of mocked get user info functions (call to GetUserInfo will pop from slice)
	GetUserInfoFns []func(userId string) (*s.User, error)
	// map stores all the sent messages by channelId (key is channelId)
	OutgoingMessages map[string][]*s.OutgoingMessage
	// Slice of rich messages
	PostMessageFunc func(channel, text string, params s.PostMessageParameters) (string, string, error)
}

func NewMockClient(t *testing.T) *MockClient {

	m := &MockClient{t: t}
	m.JoinChannelFns = []func(channelId string) error{}
	m.LeaveChannelFns = []func(channelId string) (bool, error){}
	m.GetChannelInfoFns = []func(channelId string) (*s.Channel, error){}
	m.GetUserInfoFns = []func(userId string) (*s.User, error){}
	m.OutgoingMessages = make(map[string][]*s.OutgoingMessage)
	m.PostMessageFunc = func(channel, text string, params s.PostMessageParameters) (string, string, error) {
		return "", "", nil
	}
	return m
}

// --- config methods ---

func (m *MockClient) AddMockJoinChannelCall(expectedChannelName string, returnError error) {

	x := func(channelName string) error {
		errMessage := fmt.Sprintf("JoindChannel call expected %s channelName got %s", expectedChannelName, channelName)
		require.Equal(m.t, expectedChannelName, channelName, errMessage)
		return returnError
	}
	m.JoinChannelFns = append(m.JoinChannelFns, x)
}

func (m *MockClient) AddMockLeaveChannelCall(expectedChannelId string, returnNonInChannel bool, returnError error) {

	x := func(channelId string) (bool, error) {
		errMessage := fmt.Sprintf("LeaveChannel call expected %s channelId got %s", expectedChannelId, channelId)
		require.Equal(m.t, expectedChannelId, channelId, errMessage)
		return returnNonInChannel, returnError
	}
	m.LeaveChannelFns = append(m.LeaveChannelFns, x)
}

func (m *MockClient) AddMockGetChannelInfoCall(expectedChannelId string, returnChannel *s.Channel, returnError error) {

	x := func(channelId string) (*s.Channel, error) {
		errMessage := fmt.Sprintf("GetChannelInfo call expected %s channelId got %s", expectedChannelId, channelId)
		require.Equal(m.t, expectedChannelId, channelId, errMessage)
		return returnChannel, returnError
	}
	m.GetChannelInfoFns = append(m.GetChannelInfoFns, x)
}

func (m *MockClient) AddMockGetUserInfoCall(expectedUserId string, returnUser *s.User, returnError error) {

	x := func(userId string) (*s.User, error) {
		errMessage := fmt.Sprintf("GetUserInfo call expected %s userId got %s", expectedUserId, userId)
		require.Equal(m.t, expectedUserId, userId, errMessage)
		return returnUser, returnError
	}
	m.GetUserInfoFns = append(m.GetUserInfoFns, x)
}

// --- implementation ---

func (m *MockClient) JoinChannel(channelName string) (*s.Channel, error) {

	require.NotEmpty(m.t, m.JoinChannelFns, "Unexpected (not mocked) call to join channel")

	var fn func(string) error
	fn, m.JoinChannelFns = m.JoinChannelFns[0], m.JoinChannelFns[1:]
	err := fn(channelName)
	return nil, err
}

func (m *MockClient) LeaveChannel(channelId string) (notInChannel bool, err error) {

	require.NotEmpty(m.t, m.LeaveChannelFns, "Unexpected (not mocked) call to leave channel")

	var fn func(string) (bool, error)
	fn, m.LeaveChannelFns = m.LeaveChannelFns[0], m.LeaveChannelFns[1:]
	return fn(channelId)
}

func (m *MockClient) GetChannelInfo(channelId string) (*s.Channel, error) {

	require.NotEmpty(m.t, m.GetChannelInfoFns, "Unexpected (not mocked) call to get channel info")

	var fn func(string) (*s.Channel, error)
	fn, m.GetChannelInfoFns = m.GetChannelInfoFns[0], m.GetChannelInfoFns[1:]
	return fn(channelId)
}

func (m *MockClient) GetUserInfo(userId string) (*s.User, error) {

	require.NotEmpty(m.t, m.GetUserInfoFns, "Unexpected (not mocked) call to get user info")

	var fn func(string) (*s.User, error)
	fn, m.GetUserInfoFns = m.GetUserInfoFns[0], m.GetUserInfoFns[1:]
	return fn(userId)
}

func (m *MockClient) NewOutgoingMessage(message, channelId string) *s.OutgoingMessage {

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

func NewInMemoryBackup() Backup {
	return &inMemoryBackup{}
}
