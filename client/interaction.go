package client

import (
	"encoding/json"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
	"github.com/slack-go/slack"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// InteractionHandler handles interactive message response.
type InteractionHandler struct {
	slackClient       *slack.Client
	VerificationToken string
}

func (h InteractionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logger.Debugf("message action recieved ****  %v", r)

	if r.Method != http.MethodPost {
		logger.Debugf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)

	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Debugf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		logger.Debugf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		logger.Debugf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("message action recieved %v", message)
	// Only accept message from slack with valid token
	if message.Token != h.VerificationToken {
		logger.Debugf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	originalMessage := message.OriginalMessage
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&originalMessage)
	log.Printf("message action recieved %v", message)
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	hanldeButtonClickEvent(message)

}

func hanldeButtonClickEvent(message slack.InteractionCallback) flyte.Event {
	v := message
	logger.Debugf("received message=%v in channel=%s", v.OriginalMessage, v.Channel)
	logger.Debugf("Calling Flyte Event...")
	return toFlyteButtonActionEvent(message, &v.User)

}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

}

func toFlyteButtonActionEvent(event slack.InteractionCallback, user *slack.User) flyte.Event {

	return flyte.Event{
		EventDef: flyte.EventDef{Name: "ReceivedButtonAction"},
		Payload:  newButtonActionEvent(event, user),
	}
}

type newButtonAction struct {
	ChannelId       string `json:"channelId"`
	User            user   `json:"user"`
	Action          string `json:"action"`
	Timestamp       string `json:"timestamp"`
	ActionTimestamp string `json:"actionTimestamp"`
	OriginalMessage string `json:"originalMessage"`
}

func newButtonActionEvent(e slack.InteractionCallback, u *slack.User) newButtonAction {
	return newButtonAction{
		ChannelId:       e.Channel.ID,
		User:            newUser(u),
		Action:          e.ActionCallback.AttachmentActions[0].Name,
		Timestamp:       e.MessageTs,
		ActionTimestamp: e.ActionTs,
		OriginalMessage: e.OriginalMessage.Text,
	}
}
