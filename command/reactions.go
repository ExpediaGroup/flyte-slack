package command

import (
	"encoding/json"
	"fmt"
	"github.com/ExpediaGroup/flyte-client/flyte"
	"github.com/ExpediaGroup/flyte-slack/client"
	"strings"
)

var (
	getReactionMessageInfoEventDef       = flyte.EventDef{Name: "GetReactionMessageInfoSuccess"}
	getReactionMessageInfoFailedEventDef = flyte.EventDef{Name: "GetReactionMessageInfoFailed"}
)

type GetReactionMessageInfoInput struct {
	Count           int    `json:"count"`
	Message         string `json:"message"`
	ThreadTimestamp string `json:"threadTimestamp"`
	User            string `json:"reactionUser"`
	ChannelId       string `json:"channelId"`
	ItemUser        string `json:"itemUser"`
}

type GetReactionMessageInfoOutput struct {
	GetReactionMessageInfoInput
}

type GetReactionMessageInfoFailed struct {
	GetReactionMessageInfoOutput
	Error string `json:"error"`
}

func GetReactionMessageInfo(slack client.Slack) flyte.Command {

	return flyte.Command{
		Name:         "GetReactionMessageInfo",
		OutputEvents: []flyte.EventDef{getReactionMessageInfoEventDef, getReactionMessageInfoFailedEventDef},
		Handler:      getReactionMessageInfoHandler(slack),
	}
}

func getReactionMessageInfoHandler(slack client.Slack) func(json.RawMessage) flyte.Event {

	return func(rawInput json.RawMessage) flyte.Event {
		input := GetReactionMessageInfoInput{}
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
			return getReactionMessageFailedEvent(input.Message, input.ChannelId, strings.Join(errorMessages, ", "))
		}

		issueSummary, err := slack.GetReactionMessageText(input.Count, input.User, input.ChannelId, input.ThreadTimestamp)
		if err != nil {
			resp := fmt.Errorf("Got the error = [%s] while listing the reaction for user %s Channel= %s ", err, input.User, input.ChannelId)
			errorMessages = append(errorMessages, fmt.Sprintf("input is not valid: %v", resp))
			return getReactionMessageFailedEvent(input.Message, input.ChannelId, strings.Join(errorMessages, ", "))
		}

		return getReactionMessageSuccessInfoEvent(issueSummary, input.ChannelId)
	}
}

func getReactionMessageSuccessInfoEvent(message, channelId string) flyte.Event {

	return flyte.Event{
		EventDef: getReactionMessageInfoEventDef,
		Payload:  GetReactionMessageInfoOutput{GetReactionMessageInfoInput: GetReactionMessageInfoInput{Message: message, ChannelId: channelId}},
	}
}

func getReactionMessageFailedEvent(message, channelId string, err string) flyte.Event {

	output := GetReactionMessageInfoOutput{GetReactionMessageInfoInput{Message: message, ChannelId: channelId}}
	return flyte.Event{
		EventDef: getReactionMessageInfoFailedEventDef,
		Payload:  GetReactionMessageInfoFailed{GetReactionMessageInfoOutput: output, Error: err},
	}
}
