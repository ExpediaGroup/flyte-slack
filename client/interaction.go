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
	slackClient         *slack.Client
	VerificationToken   string
	InteractionMessages flyte.Pack
}

func (h InteractionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logger.Debugf("message action recieved from slack payload= %+v", r)

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
	logger.Debugf(" Success to unespace request body: %s", jsonStr)

	var message slack.InteractionCallback

	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		logger.Debugf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	logger.Debugf("\nSuccess to decode json message from slack: %s\n", jsonStr)
	empJSON, err := json.Marshal(message)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf(string(empJSON))

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
	//TODO add a switch case for close and edit events , rightnow this is generic
	v := message
	e := toFlyteButtonActionEvent(message, &v.User)
	h.InteractionMessages.SendEvent(e)

}
