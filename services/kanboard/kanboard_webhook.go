package kanboard

import (
	"net/http"
	"strconv"

	"github.com/matrix-org/go-neb/services/kanboard/webhook"
	"github.com/matrix-org/go-neb/types"
	log "github.com/sirupsen/logrus"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const WebhookServiceType = "kanboard-webhook"

type WebhookService struct {
	types.DefaultService
	webhookEndpointURL string
	Endpoint           string `json:"endpoint"`
	Rooms              map[id.RoomID]struct {
		Projects map[string]struct {
			Events []string
		}
	}
	// Optional. The secret token to supply when creating the webhook. If supplied,
	// Go-NEB will perform security checks on incoming webhook requests using this token.
	SecretToken string
}

func (s *WebhookService) OnReceiveWebhook(w http.ResponseWriter, req *http.Request, cli types.MatrixClient) {
	eventName, projectID, msg, err := webhook.OnReceiveRequest(req, s.SecretToken, s.Endpoint)
	if err != nil {
		log.WithError(err).Error("Failed to parse webhook request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger := log.WithFields(log.Fields{
		"event":   eventName,
		"project": projectID,
	})
	for roomID, roomConfig := range s.Rooms {
		for cfgProjectID, projectConfig := range roomConfig.Projects {
			cfgProjectIDint, err := strconv.Atoi(cfgProjectID)
			if err != nil {
				logger.WithError(err).Error("Failed to parse project ID")
				continue
			}
			if projectID != cfgProjectIDint {
				continue
			}
			notifyRoom := false
			for _, cfgEventName := range projectConfig.Events {
				if cfgEventName == eventName {
					notifyRoom = true
					break
				}
			}
			if notifyRoom {
				logger.WithFields(log.Fields{
					"message": msg,
					"room":    roomID,
				}).Print("Sending notification to room")
				if _, err := cli.SendMessageEvent(roomID, event.EventMessage, msg); err != nil {
					logger.WithError(err).WithField("room", roomID).Print("Failed to send notification to room.")
				}
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}

func init() {
	types.RegisterService(func(serviceID string, serviceUserID id.UserID, webhookEndpointURL string) types.Service {
		return &WebhookService{
			DefaultService:     types.NewDefaultService(serviceID, serviceUserID, WebhookServiceType),
			webhookEndpointURL: webhookEndpointURL,
		}
	})
}
