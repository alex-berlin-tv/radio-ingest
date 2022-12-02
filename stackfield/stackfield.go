package stackfield

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// A Stackfield room to which messages can be sent.
type Room struct {
	URL string
}

// Returns a new instance of [Room].
func NewRoom(url string) Room {
	return Room{
		URL: url,
	}
}

// Send a message to the room.
func (r Room) Send(msg string) error {
	var bodyDt = make(map[string]string)
	bodyDt["Title"] = msg
	bodyBt, err := json.Marshal(bodyDt)
	if err != nil {
		return err
	}
	fmt.Println(string(bodyBt))
	bodyRd := bytes.NewReader(bodyBt)
	req, err := http.NewRequest(http.MethodPost, r.URL, bodyRd)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	clt := http.Client{
		Timeout: time.Second * 5,
	}
	rsp, err := clt.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	logrus.Debugf("got stackfield response, %s", body)
	return nil
}
