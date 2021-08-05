package command

import (
	"encoding/json"
	"github.com/ExpediaGroup/flyte-slack/cache"
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/ExpediaGroup/flyte-slack/types"
	"github.com/HotelsDotCom/flyte-client/flyte"
)

var (
	getChannelInfoSuccessEventDef = flyte.EventDef{Name: "GetChannelInfoSuccess"}
	getChannelInfoSuccessFailDef  = flyte.EventDef{Name: "GetChannelInfoFail"}
)

type GetChannelInfoInput struct {
	ChannelName string `json:"channelName"`
}

type GetChannelInfoSuccess struct {
	GetChannelInfoInput
	Conversation *types.Conversation `json:"conversation,omitempty"`
}

type GetChannelInfoFail struct {
	GetChannelInfoInput
	Reason string `json:"reason"`
}

func GetChannelInfo(slack client.Slack, cache cache.Cache) flyte.Command {
	return flyte.Command{
		Name:         "GetChannelInfo",
		OutputEvents: []flyte.EventDef{messageSentEventDef, sendMessageFailedEventDef},
		Handler:      getChannelInfoHandler(slack, cache),
	}
}

func getChannelInfoHandler(slack client.Slack, cache cache.Cache) func(json.RawMessage) flyte.Event {
	return func(rawInput json.RawMessage) flyte.Event {
		input := GetChannelInfoInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return newGetChannelInfoFail(input, err.Error())
		}

		conversation, err := cache.GetChannelID(input.ChannelName, slack)
		if err != nil {
			return newGetChannelInfoFail(input, err.Error())
		}

		return newGetChannelInfoSuccess(input, conversation)
	}
}

func newGetChannelInfoSuccess(input GetChannelInfoInput, result *types.Conversation) flyte.Event {
	return flyte.Event{
		EventDef: getChannelInfoSuccessEventDef,
		Payload: GetChannelInfoSuccess{
			GetChannelInfoInput: input,
			Conversation:        result,
		},
	}
}

func newGetChannelInfoFail(input GetChannelInfoInput, reason string) flyte.Event {
	return flyte.Event{
		EventDef: getChannelInfoSuccessFailDef,
		Payload: GetChannelInfoFail{
			GetChannelInfoInput: input,
			Reason:              reason,
		},
	}
}
