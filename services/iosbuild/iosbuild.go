// Package iosbuild implements a Service which adds !commands for iosbuild command.
package iosbuild

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/matrix-org/go-neb/types"
	mevt "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// ServiceType of the iosbuild service
const ServiceType = "iosbuild"
const hookListener = "http://office.globekeeper.com:5549"

// Service contains the Config fields for the iosbuild service.
type Service struct {
	types.DefaultService
}

// Commands supported:
//
//	!iosbuild project_name version_number
//
// Invokes hookListener `/` on mac mini.
func (s *Service) Commands(client types.MatrixClient) []types.Command {
	return []types.Command{
		{
			Path: []string{"iosbuild"},
			Command: func(roomID id.RoomID, userID id.UserID, args []string) (interface{}, error) {
				return s.cmdIosbuild(client, roomID, userID, args)
			},
		},
	}
}

// usageMessage returns a matrix TextMessage representation of the service usage
func usageMessage() *mevt.MessageEventContent {
	return &mevt.MessageEventContent{
		MsgType: mevt.MsgNotice,
		Body:    "Usage: !iosbuild project_name version_number",
	}
}

func (s *Service) cmdIosbuild(client types.MatrixClient, roomID id.RoomID, userID id.UserID, args []string) (interface{}, error) {
	// [community || globekeeper || connect, version number]
	if len(args) < 1 {
		return usageMessage(), nil
	}

	// Make the request to the iosbuild endpoint
	jsonData := map[string]string{"text": strings.Join(args, " ")}
	payload, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %s, for json: %v", err.Error(), jsonData)
	}

	req, err := http.NewRequest("POST", hookListener, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("Failed to create new request to hookListener")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to make request to hookListener: %d, %s", res.StatusCode, response2String(res))
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return mevt.MessageEventContent{
			MsgType: "m.notice",
			Body:    fmt.Sprintf("Failed to make request to hookListener: %v", response2String(res)),
		}, nil
	}

	message := "Building connect iOS..."
	if len(args) > 1 {
		message = fmt.Sprintf("Building connect iOS v%s...", args[1])
	}

	// Return article extract
	return mevt.MessageEventContent{
		MsgType: "m.notice",
		Body:    message,
	}, nil
}

// response2String returns a string representation of an HTTP response body
func response2String(res *http.Response) string {
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return "Failed to decode response body"
	}
	str := string(bs)
	return str
}

// Initialise the service
func init() {
	types.RegisterService(func(serviceID string, serviceUserID id.UserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
