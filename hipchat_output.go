// Copyright 2013, Christian Vozar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type HipchatOutput struct {
	conf   *HipchatOutputConfig
	url    string
	format string
}

// Hipchat Output config struct
type HipchatOutputConfig struct {
	// Outputs the payload attribute in the HipChat message vs a full JSON message dump
	PayloadOnly bool `toml:"payload_only"`
	// Url for HttpInput to GET.
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
	// Background color for message.
	// One of "yellow", "red", "green", "purple", "gray", or "random".
	// Default is gray
	Color string
}

func (ho *HipchatOutput) ConfigStruct() interface{} {
	return &HipchatOutputConfig{
		PayloadOnly: true,
		From:        "Heka",
		Notify:      false,
		Color:       "gray",
	}
}

func (ho *HipchatOutput) sendMessage(mc string) error {
	messageUri := fmt.Sprintf("%s/rooms/message?auth_token=%s", ho.url, url.QueryEscape(ho.conf.AuthToken))

	messagePayload := url.Values{
		"room_id":        {ho.conf.RoomId},
		"from":           {ho.conf.From},
		"message":        {mc},
		"color":          {ho.conf.Color},
		"message_format": {ho.format},
	}

	if ho.conf.Notify == true {
		messagePayload.Add("notify", "1")
	}

	resp, err := http.PostForm(messageUri, messagePayload)
	if err != nil {
		return err
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
		return errors.New("HipchatOutput: Status response was not sent.")
	}

	return nil
}

func (ho *HipchatOutput) Init(config interface{}) (err error) {
	ho.conf = config.(*HipchatOutputConfig)

	if ho.conf.RoomId == nil {
		return fmt.Errorf("room_id must contain a HipChat room ID or name")
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
			err = ho.sendMessage(msg.GetPayload())
		} else {
			if contents, err = json.Marshal(msg); err == nil {
				err = ho.sendMessage(string(contents))
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
