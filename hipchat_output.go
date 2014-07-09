// Copyright 2014, Christian Vozar
// Licensed under the MIT License.

// Code based on Atlassian HipChat API v1.

package hipchat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"io/ioutil"
	"net/http"
	"net/url"
)

// HipchatOutput maintains high-level configuration options for the plugin.
type HipchatOutput struct {
	conf   *HipchatOutputConfig
	url    string
	format string
}

// Hipchat Output config struct
type HipchatOutputConfig struct {
	// Outputs the payload attribute in the HipChat message vs a full JSON message dump
	PayloadOnly bool `toml:"payload_only"`
	// HipChat Authorization token. Notification token is appropriate.
	AuthToken string `toml:"auth_token"`
	// Required. ID or name of the room.
	RoomId string `toml:"room_id"`
	// Required. Name the message will appear be sent. Must be less than 15
	// characters long. May contain letters, numbers, -, _, and spaces.
	From string
	// Whether or not this message should trigger a notification for people
	// in the room (change the tab color, play a sound, etc).
	// Each recipient's notification preferences are taken into account.
	// Default is false
	Notify bool
}

func (ho *HipchatOutput) ConfigStruct() interface{} {
	return &HipchatOutputConfig{
		PayloadOnly: true,
		From:        "Heka",
		Notify:      false,
	}
}

func (ho *HipchatOutput) sendMessage(mc string, s int32) error {
	messageUri := fmt.Sprintf("%s/rooms/message?auth_token=%s", ho.url, url.QueryEscape(ho.conf.AuthToken))

	messagePayload := url.Values{
		"room_id":        {ho.conf.RoomId},
		"from":           {ho.conf.From},
		"message":        {mc},
		"message_format": {ho.format},
	}

	if ho.conf.Notify == true {
		messagePayload.Add("notify", "1")
	}

	switch s {
	case 0, 1, 2, 3:
		messagePayload.Add("color", "red")
	case 4:
		messagePayload.Add("color", "yellow")
	case 5, 6:
		messagePayload.Add("color", "green")
	default:
		messagePayload.Add("color", "gray")
	}

	resp, err := http.PostForm(messageUri, messagePayload)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 400:
		return errors.New("Bad request.")
	case 401:
		return errors.New("Provided authentication rejected.")
	case 403:
		return errors.New("Rate limit exceeded.")
	case 406:
		return errors.New("Message contains invalid content type.")
	case 500:
		return errors.New("Internal server error.")
	case 503:
		return errors.New("Service unavailable.")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	messageResponse := &struct{ Status string }{}
	if err := json.Unmarshal(body, messageResponse); err != nil {
		return err
	}
	if messageResponse.Status != "sent" {
		return errors.New("Status response was not sent.")
	}

	return nil
}

func (ho *HipchatOutput) Init(config interface{}) (err error) {
	ho.conf = config.(*HipchatOutputConfig)

	if ho.conf.RoomId == "" {
		return fmt.Errorf("room_id must contain a HipChat room ID or name.")
	}

	if len(ho.conf.From) > 15 {
		return fmt.Errorf("from must be less than 15 characters.")
	}

	ho.url = "https://api.hipchat.com/v1"
	ho.format = "text"
	return
}

func (ho *HipchatOutput) Run(or OutputRunner, h PluginHelper) (err error) {
	inChan := or.InChan()

	var (
		pack     *PipelinePack
		msg      *message.Message
		contents []byte
	)

	for pack = range inChan {
		msg = pack.Message
		if ho.conf.PayloadOnly {
			err = ho.sendMessage(msg.GetPayload(), msg.GetSeverity())
		} else {
			if contents, err = json.Marshal(msg); err == nil {
				err = ho.sendMessage(string(contents), msg.GetSeverity())
			} else {
				or.LogError(err)
			}
		}
		if err != nil {
			or.LogError(err)
		}
		pack.Recycle()
	}
	return
}

func init() {
	RegisterPlugin("HipchatOutput", func() interface{} {
		return new(HipchatOutput)
	})
}
