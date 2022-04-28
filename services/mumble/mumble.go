package mumble

import (
	"crypto/tls"
	"fmt"
	"net"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"

	"github.com/matrix-org/go-neb/types"
	mevt "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const ServiceType = "mumble"

type Service struct {
	types.DefaultService
	Endpoint     string   `json:"endpoint"`
	Insecure     bool     `json:"insecure"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Room         string   `json:"room"`
	IgnoredUsers []string `json:"ignoredUsers"`
}

func (s *Service) Register(oldService types.Service, client types.MatrixClient) error {

	config := gumble.NewConfig()
	config.Username = s.Username
	config.Password = s.Password

	config.Attach(gumbleutil.Listener{
		UserChange: func(e *gumble.UserChangeEvent) {
			for _, u := range s.IgnoredUsers {
				if u == e.User.Name {
					return
				}
			}

			if e.Type.Has(gumble.UserChangeConnected) {
				msg := mevt.MessageEventContent{
					Body:    fmt.Sprintf("User %s has joined Mumble", e.User.Name),
					MsgType: "m.notice",
				}
				client.SendMessageEvent(id.RoomID(s.Room), mevt.EventMessage, msg)
			} else if e.Type.Has(gumble.UserChangeDisconnected) {
				msg := mevt.MessageEventContent{
					Body:    fmt.Sprintf("User %s has left Mumble", e.User.Name),
					MsgType: "m.notice",
				}
				client.SendMessageEvent(id.RoomID(s.Room), mevt.EventMessage, msg)
			}
		},
	})
	var tlsConfig tls.Config
	if s.Insecure {
		tlsConfig = tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		tlsConfig = tls.Config{}
	}

	_, err := gumble.DialWithDialer(new(net.Dialer), s.Endpoint, config, &tlsConfig)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	types.RegisterService(func(serviceID string, serviceUserID id.UserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
