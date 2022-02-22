package command

import (
	"encoding/json"
	"fmt"
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"strings"
)

var (
	getReactionListEventDef       = flyte.EventDef{Name: "GetReactionListSuccess"}
	getReactionListFailedEventDef = flyte.EventDef{Name: "GetReactionListFailed"}
)

type GetReactionListInput struct {
	Count           int    `json:"count"`
	Message         string `json:"message"`
	ThreadTimestamp string `json:"threadTimestamp"`
	User            string `json:"reactionUser"`
	ChannelId       string `json:"channelId"`
	ItemUser        string `json:"itemUser"`
}

type GetReactionListOutput struct {
	GetReactionListInput
}

type GetReactionListErrorOutput struct {
	GetReactionListOutput
	Error string `json:"error"`
}

func GetReactionList(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "GetReactionList",
		OutputEvents: []flyte.EventDef{getReactionListEventDef, getReactionListFailedEventDef},
		Handler:      getReactionListHandler(slack),
	}
}

func getReactionListHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {
		input := GetReactionListInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("input is not valid: %v", err))
		}

		errorMessages := []string{}
		if input.ThreadTimestamp == "" {
			errorMessages = append(errorMessages, "missing Message Timestamp field")
		}
		if input.ChannelId == "" {
			errorMessages = append(errorMessages, "missing channel id field")
		}
		if input.User == "" {
			errorMessages = append(errorMessages, "missing user id field")
		}
		if len(errorMessages) != 0 {
			return newreactionListFailedEvent(input.Message, input.ChannelId, strings.Join(errorMessages, ", "))
		}

		issueSummary := slack.ListReactions(50, input.User, input.ChannelId, input.ThreadTimestamp)
		return newreactionListEvent(issueSummary, input.ChannelId)
	}
}

func newreactionListEvent(message, channelId string) flyte.Event {

	return flyte.Event{
		EventDef: getReactionListEventDef,
		Payload:  GetReactionListOutput{GetReactionListInput: GetReactionListInput{Message: message, ChannelId: channelId}},
	}
}

func newreactionListFailedEvent(message, channelId string, err string) flyte.Event {

	output := GetReactionListOutput{GetReactionListInput: GetReactionListInput{Message: message, ChannelId: channelId}}
	return flyte.Event{
		EventDef: getReactionListFailedEventDef,
		Payload:  GetReactionListErrorOutput{GetReactionListOutput: output, Error: err},
	}
}
