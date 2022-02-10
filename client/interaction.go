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

type interactionclient interface {
	GetUserInfo(userId string) (*slack.User, error)
	NewOutgoingMessage(message, channelId string, options ...slack.RTMsgOption) *slack.OutgoingMessage
	SendMessage(message *slack.OutgoingMessage)
	PostMessage(channel string, opts ...slack.MsgOption) (string, string, error)
	GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
	GetReactions(item slack.ItemRef, params slack.GetReactionsParameters) (reactions []slack.ItemReaction, err error)
	ListReactions(params slack.ListReactionsParameters) ([]slack.ReactedItem, *slack.Paging, error)
}

// InteractionHandler handles interactive message response.
type InteractionHandler struct {
	slackClient       *slack.Client
	VerificationToken string
	// messages to be consumed by API (filtered incoming events)
	InteractionMessages flyte.Pack
}

func (h InteractionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logger.Debugf("message action recieved ****  %+v", r)

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
	jsonStr = `{
    "type": "interactive_message",
    "token": "V5mEO0CzPYjbQ7MfStmlNqiQ",
    "callback_id": "RCPSUP-304",
    "response_url": "https://hooks.slack.com/actions/TJ0URGP5L/3086040546290/Ik4KaRK4zC0AxwdVPdILkDxZ",
    "trigger_id": "3079379823990.612977567190.f277be21ac16a7c96c84ae447f5dd600",
    "action_ts": "1644500468.988242",
    "team": {
        "id": "TJ0URGP5L",
        "name": "",
        "domain": "expedia-sandbox"
    },
    "channel": {
        "id": "C02RJN7FTSM",
        "created": 0,
        "is_open": false,
        "is_group": false,
        "is_shared": false,
        "is_im": false,
        "is_ext_shared": false,
        "is_org_shared": false,
        "is_pending_ext_shared": false,
        "is_private": false,
        "is_mpim": false,
        "unlinked": 0,
        "name_normalized": "",
        "num_members": 0,
        "priority": 0,
        "user": "",
        "name": "flyte-cse-test",
        "creator": "",
        "is_archived": false,
        "members": null,
        "topic": {
            "value": "",
            "creator": "",
            "last_set": 0
        },
        "purpose": {
            "value": "",
            "creator": "",
            "last_set": 0
        },
        "is_channel": false,
        "is_general": false,
        "is_member": false,
        "locale": ""
    },
    "user": {
        "id": "U01NAB6ERFB",
        "team_id": "TJ0URGP5L",
        "name": "songupta",
        "deleted": false,
        "color": "",
        "real_name": "",
        "tz_label": "",
        "tz_offset": 0,
        "profile": {
            "first_name": "",
            "last_name": "",
            "real_name": "",
            "real_name_normalized": "",
            "display_name": "",
            "display_name_normalized": "",
            "email": "",
            "skype": "",
            "phone": "",
            "image_24": "",
            "image_32": "",
            "image_48": "",
            "image_72": "",
            "image_192": "",
            "image_512": "",
            "image_original": "",
            "title": "",
            "status_expiration": 0,
            "team": "",
            "fields": []
        },
        "is_bot": false,
        "is_admin": false,
        "is_owner": false,
        "is_primary_owner": false,
        "is_restricted": false,
        "is_ultra_restricted": false,
        "is_stranger": false,
        "is_app_user": false,
        "is_invited_user": false,
        "has_2fa": false,
        "has_files": false,
        "presence": "",
        "locale": "",
        "updated": 0,
        "enterprise_user": {
            "id": "",
            "enterprise_id": "",
            "enterprise_name": "",
            "is_admin": false,
            "is_owner": false,
            "teams": null
        }
    },
    "original_message": {
        "type": "message",
        "user": "U02TP6M5QGM",
        "ts": "1644500452.940239",
        "thread_ts": "1644500363.725629",
        "attachments": [
            {
                "color": "36a64f",
                "fallback": "RCPSUP_FALLBACK",
                "callback_id": "RCPSUP-304",
                "id": 1,
                "title": "JIRA issue created successfully! | \u003chttps://jira.expedia.biz/browse/RCPSUP-304|RCPSUP-304\u003e",
                "fields": [
                    {
                        "title": "Summary",
                        "value": "Facing issue sample test...",
                        "short": true
                    },
                    {
                        "title": "Status",
                        "value": "Open",
                        "short": true
                    },
                    {
                        "title": "Project",
                        "value": "RCPSUP",
                        "short": true
                    },
                    {
                        "title": "Priority",
                        "value": "Medium",
                        "short": true
                    }
                ],
                "actions": [
                    {
                        "name": "Jira ticket closed",
                        "text": "Close Issue",
                        "style": "danger",
                        "type": "button",
                        "value": "RCPSUP-304",
                        "confirm": {
                            "title": "Are you sure?",
                            "text": "Are you sure?",
                            "ok_text": "Yes",
                            "dismiss_text": "No"
                        }
                    }
                ],
                "blocks": null
            }
        ],
        "bot_id": "B02TBFBMQQP",
        "bot_profile": {
            "app_id": "A02TBEBMEB1",
            "icons": {
                "image_36": "https://a.slack-edge.com/80588/img/plugins/app/bot_36.png",
                "image_48": "https://a.slack-edge.com/80588/img/plugins/app/bot_48.png",
                "image_72": "https://a.slack-edge.com/80588/img/plugins/app/service_72.png"
            },
            "id": "B02TBFBMQQP",
            "name": "rcpbot",
            "team_id": "TJ0URGP5L",
            "updated": 1642014043
        },
        "parent_user_id": "U01NAB6ERFB",
        "team": "TJ0URGP5L",
        "replace_original": false,
        "delete_original": false,
        "blocks": null
    },
    "message": {
        "replace_original": false,
        "delete_original": false,
        "blocks": null
    },
    "name": "",
    "value": "",
    "message_ts": "1644500452.940239",
    "attachment_id": "1",
    "actions": [
        {
            "name": "Jira ticket closed",
            "text": "",
            "type": "button",
            "value": "RCPSUP-304"
        }
    ],
    "view": {
        "ok": false,
        "error": "",
        "response_metadata": {
            "next_cursor": "",
            "messages": null,
            "warnings": null
        },
        "id": "",
        "team_id": "",
        "type": "",
        "title": null,
        "close": null,
        "submit": null,
        "blocks": null,
        "private_metadata": "",
        "callback_id": "",
        "state": null,
        "hash": "",
        "clear_on_close": false,
        "notify_on_close": false,
        "root_view_id": "",
        "previous_view_id": "",
        "app_id": "",
        "external_id": "",
        "bot_id": ""
    },
    "action_id": "",
    "api_app_id": "",
    "block_id": "",
    "container": {
        "type": "",
        "view_id": "",
        "message_ts": "",
        "attachment_id": 0,
        "channel_id": "",
        "is_ephemeral": false,
        "is_app_unfurl": false
    },
    "submission": null,
    "hash": "",
    "is_cleared": false
}`

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
	//hanldeButtonClickEvent(message)
	v := message

	e := toFlyteButtonActionEvent(message, &v.User)
	h.InteractionMessages.SendEvent(e)

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
