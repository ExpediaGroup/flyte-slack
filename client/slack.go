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
	s "github.com/nlopes/slack"
	"github.com/HotelsDotCom/go-logger"
	"sync"
	"errors"
	"fmt"
	"github.com/HotelsDotCom/flyte-client/flyte"
)

// slack RTM uses channel name to join channel, but everything else uses channel id
// therefore we need to call GetChannelInfo to retrieve channel name when joining
type client interface {
	JoinChannel(channelName string) (*s.Channel, error)
	LeaveChannel(channelId string) (notInChannel bool, err error)
	GetChannelInfo(channelId string) (*s.Channel, error)
	GetUserInfo(userId string) (*s.User, error)
	NewOutgoingMessage(message, channelId string) *s.OutgoingMessage
	SendMessage(message *s.OutgoingMessage)
	PostMessage(channel, text string, params s.PostMessageParameters) (string, string, error)
}

// our slack implementation makes consistent use of channel id
type Slack interface {
	SendMessage(message, channelId string)
	SendRichMessage(rm RichMessage) error
	Broadcast(message string)
	JoinChannel(channelId string)
	LeaveChannel(channelId string)
	JoinedChannels() []string
	IncomingMessages() <-chan flyte.Event
}

type slack struct {
	backup Backup
	client client
	// events received from slack
	incomingEvents chan s.RTMEvent
	// messages to be consumed by API (filtered incoming events)
	incomingMessages chan flyte.Event
	joinedChannelIds *sync.Map
}

// Creates new slack client with supplied backup, slack token - https://api.slack.com/custom-integrations/legacy-tokens
// and default channel to join. Default channel can be empty if backup contains at least one channel. Basically client
// will fail if it cannot join any channel (either from backup or provided as deafultChannel argument)
func NewSlack(backup Backup, token string, defaultChannel string) Slack {

	// load all the channels we need to join
	channelIds := backup.Load()
	if defaultChannel != "" {
		channelIds = append(channelIds, defaultChannel)
	}

	// initialize slack client
	slackRtm := s.New(token).NewRTM()
	go slackRtm.ManageConnection()

	sl := &slack{
		backup:           backup,
		client:           slackRtm,
		incomingEvents:   slackRtm.IncomingEvents,
		incomingMessages: make(chan flyte.Event),
		joinedChannelIds: &sync.Map{},
	}

	logger.Info("initialized slack")
	if len(channelIds) == 0 {
		logger.Info("no channels joined")
	}

	// join channels and listen for incoming messages
	for _, id := range channelIds {
		sl.JoinChannel(id)
	}
	go sl.handleMessageEvents()
	return sl
}

// Sends slack message to provided channel. Channel does not have to be joined.
func (sl *slack) SendMessage(message, channelId string) {

	msg := sl.client.NewOutgoingMessage(message, channelId)
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

// Broadcast message - sends message to all joined channels.
func (sl *slack) Broadcast(message string) {

	sl.joinedChannelIds.Range(func(channelId, _ interface{}) bool {
		sl.SendMessage(message, channelId.(string))
		return true
	})
}

// Join channel, all the received messages on this channel will be available in IncomingMessages() channel and calling
// Broadcast() will send messages to this channel.
func (sl *slack) JoinChannel(channelId string) {

	if _, ok := sl.joinedChannelIds.LoadOrStore(channelId, true); ok {
		return
	}

	channel, err := sl.client.GetChannelInfo(channelId)
	if err != nil {
		logger.Errorf("cannot join channel=%s: cannot get channel info: %v", channelId, err)
		return
	}

	if _, err = sl.client.JoinChannel(channel.Name); err != nil {
		logger.Errorf("cannot join channel=%s: %v", channelId, err)
		return
	}

	sl.backup.Backup(sl.JoinedChannels())
	sl.SendMessage("Hello! I've joined this room ...", channelId)
	logger.Infof("joined channel=%s", channelId)
}

// Leaves channel, messages received in this channel will not be available in IncomingMessages() channel and Broadcast
// will not send messages to this channel.
func (sl *slack) LeaveChannel(channelId string) {

	// rtm does not send message if channel is not joined
	sl.SendMessage("I'm leaving now, bye!", channelId)

	if _, err := sl.client.LeaveChannel(channelId); err != nil {
		logger.Errorf("cannot leave channel=%s: %v", channelId, err)
		return
	}

	sl.joinedChannelIds.Delete(channelId)
	sl.backup.Backup(sl.JoinedChannels())
	logger.Infof("left channel=%s", channelId)
}

// Returns ids of all channels that are currently joined.
func (sl *slack) JoinedChannels() []string {

	channelIds := []string{}
	sl.joinedChannelIds.Range(func(key, _ interface{}) bool {
		channelIds = append(channelIds, key.(string))
		return true
	})
	return channelIds
}

// Returns channel with incoming messages from all joined channels.
func (sl *slack) IncomingMessages() <-chan flyte.Event {
	return sl.incomingMessages
}

func (sl *slack) handleMessageEvents() {

	for event := range sl.incomingEvents {
		switch v := event.Data.(type) {
		case *s.MessageEvent:
			if _, ok := sl.joinedChannelIds.Load(v.Channel); ok {
				logger.Infof("received message=%s in channel=%s", v.Text, v.Channel)
				u, err := sl.client.GetUserInfo(v.User)
				if err != nil {
					logger.Errorf("cannot get info about user=%s: %v", v.User, err)
					continue
				}
				sl.incomingMessages <- toFlyteMessageEvent(v, u)
			}
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
	ChannelId string `json:"channelId"`
	User      user   `json:"user"`
	Message   string `json:"message"`
}

func newMessageEvent(e *s.MessageEvent, u *s.User) messageEvent {

	return messageEvent{
		ChannelId: e.Channel,
		User:      newUser(u),
		Message:   e.Text,
	}
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
