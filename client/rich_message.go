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

import "github.com/slack-go/slack"

type RichMessage struct {
	Parse           string             `json:"parse"`
	ThreadTimestamp string             `json:"thread_ts"`
	ReplyBroadcast  bool               `json:"reply_broadcast"`
	LinkNames       int                `json:"link_names"`
	Attachments     []slack.Attachment `json:"attachments"`
	UnfurlLinks     bool               `json:"unfurl_links"`
	UnfurlMedia     bool               `json:"unfurl_media"`
	IconURL         string             `json:"icon_url"`
	IconEmoji       string             `json:"icon_emoji"`
	Markdown        bool               `json:"mrkdwn,omitempty"`
	EscapeText      bool               `json:"escape_text"`
	ChannelID       string             `json:"channel"`
	Text            string             `json:"text"`
}

type MessagePoster interface {
	PostMessage(channel string, params ...slack.MsgOption) (string, string, error)
}

func (m RichMessage) Post(rtm MessagePoster) error {
	_, _, err := rtm.PostMessage(m.ChannelID, m.toMsgOptions()...)
	return err
}

func (m RichMessage) toMsgOptions() []slack.MsgOption {
	return []slack.MsgOption{
		slack.MsgOptionText(m.Text, m.EscapeText),
		slack.MsgOptionPostMessageParameters(m.toPostMessageParameters()),
		slack.MsgOptionAttachments(m.Attachments...),
	}
}

func (m RichMessage) toPostMessageParameters() slack.PostMessageParameters {
	return slack.PostMessageParameters{
		AsUser:          true,
		Parse:           m.Parse,
		ThreadTimestamp: m.ThreadTimestamp,
		ReplyBroadcast:  m.ReplyBroadcast,
		LinkNames:       m.LinkNames,
		UnfurlLinks:     m.UnfurlLinks,
		UnfurlMedia:     m.UnfurlMedia,
		IconURL:         m.IconURL,
		IconEmoji:       m.IconEmoji,
		Markdown:        m.Markdown,
		EscapeText:      m.EscapeText,
	}
}
