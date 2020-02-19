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
	s "github.com/nlopes/slack"
)

type client interface {
	GetUserInfo(userId string) (*s.User, error)
	NewOutgoingMessage(message, channelId string, options ...s.RTMsgOption) *s.OutgoingMessage
	SendMessage(message *s.OutgoingMessage)
	PostMessage(channel, text string, params s.PostMessageParameters) (string, string, error)
}

// our slack implementation makes consistent use of channel id
type Slack interface {
	SendMessage(message, channelId, threadTimestamp string)
	SendRichMessage(rm RichMessage) error
	IncomingMessages() <-chan flyte.Event
}

type slack struct {
	client client
	// events received from slack
	incomingEvents chan s.RTMEvent
	// messages to be consumed by API (filtered incoming events)
	incomingMessages chan flyte.Event
}

func NewSlack(token string) Slack {

	rtm := s.New(token).NewRTM()
	go rtm.ManageConnection()

	sl := &slack{
		client:           rtm,
		incomingEvents:   rtm.IncomingEvents,
		incomingMessages: make(chan flyte.Event),
	}

	logger.Info("initialized slack")
	go sl.handleMessageEvents()
	return sl
}

// Sends slack message to provided channel. Channel does not have to be joined.
func (sl *slack) SendMessage(message, channelId, threadTimestamp string) {

	msg := sl.client.NewOutgoingMessage(message, channelId)
	msg.ThreadTimestamp = threadTimestamp
	sl.client.SendMessage(msg)
	logger.Infof("message=%q sent to channel=%s", msg.Text, channelId)

}

func (sl *slack) SendRichMessage(rm RichMessage) error {
	if err := rm.Post(sl.client); err != nil {
		return errors.New(fmt.Sprintf("cannot send rich message=%v: %v", rm, err))
	}
	logger.Infof("rich message=%+v sent to channel=%s", rm, rm.ChannelID)
	return nil
}

// Returns channel with incoming messages from all joined channels.
func (sl *slack) IncomingMessages() <-chan flyte.Event {
	return sl.incomingMessages
}

func (sl *slack) handleMessageEvents() {
	for event := range sl.incomingEvents {
		switch v := event.Data.(type) {
		case *s.MessageEvent:
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

func toFlyteMessageEvent(event *s.MessageEvent, user *s.User) flyte.Event {

	return flyte.Event{
		EventDef: flyte.EventDef{Name: "ReceivedMessage"},
		Payload:  newMessageEvent(event, user),
	}
}

type messageEvent struct {
	ChannelId       string `json:"channelId"`
	User            user   `json:"user"`
	Message         string `json:"message"`
	Timestamp       string `json:"timestamp"`
	ThreadTimestamp string `json:"threadTimestamp"`
	Thread string `json:"thread"`
}

func newMessageEvent(e *s.MessageEvent, u *s.User) messageEvent {
	return messageEvent{
		ChannelId:       e.Channel,
		User:            newUser(u),
		Message:         e.Text,
		Timestamp:       e.Timestamp,
		ThreadTimestamp: e.ThreadTimestamp,
		Thread: getThread(e),
	}
}

func getThread(e *s.MessageEvent) string {
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

func newUser(u *s.User) user {

	return user{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Profile.Email,
		Title:     u.Profile.Title,
		FirstName: u.Profile.FirstName,
		LastName:  u.Profile.LastName,
	}
}
