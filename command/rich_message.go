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
	"github.com/HotelsDotCom/flyte-slack/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"encoding/json"
	"fmt"
	"github.com/HotelsDotCom/go-logger"
)

var (
	richMessageSentEventDef       = flyte.EventDef{Name: "RichMessageSent"}
	sendRichMessageFailedEventDef = flyte.EventDef{Name: "SendRichMessageFailed"}
)

type SendRichMessageErrorOutput struct {
	InputMessage client.RichMessage `json:"inputMessage"`
	Error        string             `json:"error"`
}

type RichMessageSender interface {
	SendRichMessage(rm client.RichMessage) error
}

func SendRichMessage(sender RichMessageSender) flyte.Command {
	return flyte.Command{
		Name:         "SendRichMessage",
		OutputEvents: []flyte.EventDef{richMessageSentEventDef, sendRichMessageFailedEventDef},
		Handler:      sendRichMessageHandler(sender),
	}
}

func sendRichMessageHandler(sender RichMessageSender) flyte.CommandHandler {
	return func(rawInput json.RawMessage) flyte.Event {
		var input client.RichMessage
		if err := json.Unmarshal(rawInput, &input); err != nil {
			errorMessage := fmt.Sprintf("invalid input: %v", err)
			logger.Errorf(errorMessage)
			return flyte.NewFatalEvent(errorMessage)
		}

		if err := sender.SendRichMessage(input); err != nil {
			logger.Errorf("error sending rich message: %v", err)
			return flyte.Event{
				EventDef: sendRichMessageFailedEventDef,
				Payload: SendRichMessageErrorOutput{
					InputMessage:input,
					Error:err.Error(),
				},
			}
		}

		return flyte.Event{
			EventDef: richMessageSentEventDef,
			Payload:  input,
		}
	}
}
