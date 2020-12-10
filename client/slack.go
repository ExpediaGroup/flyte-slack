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
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
	"github.com/slack-go/slack"
)

type client interface {
	GetUserInfo(userId string) (*slack.User, error)
	NewOutgoingMessage(message, channelId string, options ...slack.RTMsgOption) *slack.OutgoingMessage
	SendMessage(message *slack.OutgoingMessage)
	PostMessage(channel string, opts ...slack.MsgOption) (string, string, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
}

// our slack implementation makes consistent use of channel id
type Slack interface {
	SendMessage(message, channelId, threadTimestamp string)
	SendRichMessage(rm RichMessage) (respChannel string, respTimestamp string, err error)
	IncomingMessages() <-chan flyte.Event
	GetConversationReplies(channel string, threadTimestamp string) ([]slack.Message, error)
}

type slackClient struct {
	client client
	// events received from slack
	incomingEvents chan slack.RTMEvent
	// messages to be consumed by API (filtered incoming events)
	incomingMessages chan flyte.Event
}

func (sl *slackClient) GetConversationReplies(channelId, threadTimestamp string) ([]slack.Message, error) {
	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelId,
		Timestamp: threadTimestamp,
	}

	msg, _, _, err := sl.client.GetConversationReplies(params)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get channel replies=%v", err))
	}
	logger.Infof("received message reply for timestamp=%s sent to channel=%s", threadTimestamp, channelId)

	return msg, nil
}

func NewSlack(token string) Slack {

	rtm := slack.New(token).NewRTM()
	go rtm.ManageConnection()

	sl := &slackClient{
		client:           rtm,
		incomingEvents:   rtm.IncomingEvents,
		incomingMessages: make(chan flyte.Event),
	}

	logger.Info("initialized slack")
	go sl.handleMessageEvents()
	return sl
}

// Sends slack message to provided channel. Channel does not have to be joined.
func (sl *slackClient) SendMessage(message, channelId, threadTimestamp string) {

	msg := sl.client.NewOutgoingMessage(message, channelId)
	msg.ThreadTimestamp = threadTimestamp
	sl.client.SendMessage(msg)
	logger.Infof("message=%q sent to channel=%s", msg.Text, channelId)

}

func (sl *slackClient) SendRichMessage(rm RichMessage) (string, string, error) {
	respChannel, respTimestamp, err := rm.Post(sl.client)
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("cannot send rich message=%v: %v", rm, err))
	}
	logger.Infof("rich message=%+v sent to channel=%s", rm, rm.ChannelID)
	return respChannel, respTimestamp, nil
}

// Returns channel with incoming messages from all joined channels.
func (sl *slackClient) IncomingMessages() <-chan flyte.Event {
	return sl.incomingMessages
}

func (sl *slackClient) handleMessageEvents() {
	for event := range sl.incomingEvents {
		switch v := event.Data.(type) {
		case *slack.MessageEvent:
			logger.Debugf("received message=%s in channel=%s", v.Text, v.Channel)
			u, err := sl.client.GetUserInfo(v.User)
			if err != nil {
				logger.Errorf("cannot get info about user=%s: %v", v.User, err)
				continue
			}
			sl.incomingMessages <- toFlyteMessageEvent(v, u)
		}
	}
}

func toFlyteMessageEvent(event *slack.MessageEvent, user *slack.User) flyte.Event {

	return flyte.Event{
		EventDef: flyte.EventDef{Name: "ReceivedMessage"},
		Payload:  newMessageEvent(event, user),
	}
}

type messageEvent struct {
	ChannelId       string        `json:"channelId"`
	User            user          `json:"user"`
	Message         string        `json:"message"`
	Timestamp       string        `json:"timestamp"`
	ThreadTimestamp string        `json:"threadTimestamp"`
	ReplyCount      int           `json:"replyCount"`
	Replies         []slack.Reply `json:"replies"`
}

func newMessageEvent(e *slack.MessageEvent, u *slack.User) messageEvent {
	return messageEvent{
		ChannelId:       e.Channel,
		User:            newUser(u),
		Message:         e.Text,
		Timestamp:       e.Timestamp,
		ThreadTimestamp: getThreadTimestamp(e),
		ReplyCount:      e.ReplyCount,
		Replies:         e.Replies,
	}
}

func getThreadTimestamp(e *slack.MessageEvent) string {
	if e.ThreadTimestamp != "" {
		return e.ThreadTimestamp
	}
	return e.Timestamp
}

type user struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Title     string `json:"title"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func newUser(u *slack.User) user {

	return user{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Profile.Email,
		Title:     u.Profile.Title,
		FirstName: u.Profile.FirstName,
		LastName:  u.Profile.LastName,
	}
}
