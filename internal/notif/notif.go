package notif

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// Notifier is the interface that wraps the Notify method.
//
// Notify notifies the message to the output channel.
// The implementation channel can be slack/email/etc.
type Notifier interface {
	Notify(message string) error
}

const apiURL = "https://slack.com/api"
const defaultHTTPTimeout = 80 * time.Second

// SlackNotifierConfig represent the config needed when creating a new slack notifier
type SlackNotifierConfig struct {
	Token      string
	Channel    string
	HTTPClient *http.Client
}

// SlackNotifier represents the notifier that will notify to slack channel
type SlackNotifier struct {
	Token      string
	Channel    string
	HTTPClient *http.Client
}

// Notify notifies message to a slack channel
func (sn *SlackNotifier) Notify(message string) error {
	/*
		Examples of calling the slack API:
			curl -X POST -H 'Authorization: Bearer xoxb-1234-56789abcdefghijklmnop' \
			-H 'Content-type: application/json' \
			--data '{
				"channel":"C061EG9SL",
				"text":"I hope the tour went well, Mr. Wonka.",
				"attachments": [{
					"text":"Who wins the lifetime supply of chocolate?",
					"fallback":"You could be telling the computer exactly what it can do with a lifetime supply of chocolate.",
					"color":"#3AA3E3",
					"attachment_type":"default",
					"callback_id":"select_simple_1234",
					"actions":[{
						"name":"winners_list",
						"text":"Who should win?",
						"type":"select",
						"data_source":"users"
					}]
				}]
			}' \
			https://slack.com/api/chat.postMessage
	*/

	payload := []byte(fmt.Sprintf(`{
		"channel": "%s",
		"text": "%s",
	}`, sn.Channel, message))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat.postMessage", apiURL), bytes.NewBuffer(payload))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sn.Token))
	req.Header.Set("Content-type", "application/json")
	if err != nil {
		return err
	}

	if _, err := sn.HTTPClient.Do(req); err != nil {
		return err
	}
	return nil
}

// NewSlackNotifier creates a new slack notifier.
// If the http client is not provided, will use the default http client with default http timeout 80secs
func NewSlackNotifier(config SlackNotifierConfig) *SlackNotifier {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &SlackNotifier{
		Token:      config.Token,
		Channel:    config.Channel,
		HTTPClient: config.HTTPClient,
	}
}
