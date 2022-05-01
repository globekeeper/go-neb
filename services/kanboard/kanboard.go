package kanboard

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/elohmeier/ganboard/v2"
	"github.com/matrix-org/go-neb/types"
	mevt "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const ServiceType = "kanboard"

type Service struct {
	types.DefaultService
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (k *Service) ganboardClient() ganboard.Client {
	u, err := url.Parse(k.Endpoint)
	if err != nil {
		panic("failed to parse kanboard endpoint url")
	}
	u.Path = path.Join(u.Path, "jsonrpc.php")
	return ganboard.Client{
		Endpoint: u.String(),
		Username: k.Username,
		Password: k.Password,
	}
}

const createUsage = "Usage: !kanboard create <project_id> \"task title\""

func (k *Service) cmdKanboardCreateTask(roomID id.RoomID, userID id.UserID, args []string) (interface{}, error) {
	client := k.ganboardClient()

	if len(args) != 2 {
		return &mevt.MessageEventContent{
			MsgType: mevt.MsgNotice,
			Body:    createUsage,
		}, nil
	}

	projectID, err := strconv.Atoi(args[0])
	if err != nil {
		return &mevt.MessageEventContent{
			MsgType: mevt.MsgNotice,
			Body:    createUsage,
		}, nil
	}

	params := ganboard.TaskParams{
		Title:     args[1],
		ProjectID: projectID,
	}

	taskID, err := client.CreateTask(params)

	if err != nil {
		return &mevt.MessageEventContent{
			MsgType: mevt.MsgNotice,
			Body:    fmt.Sprint(err),
		}, nil
	}

	u, _ := url.Parse(k.Endpoint)
	u.Path = path.Join(u.Path, "task", strconv.Itoa(taskID))

	return &mevt.MessageEventContent{
		MsgType: mevt.MsgNotice,
		Body:    "Task created: " + u.String(),
	}, nil
}

func (k *Service) cmdKanboardListProjects(roomID id.RoomID, userID id.UserID, args []string) (interface{}, error) {
	client := k.ganboardClient()

	projects, err := client.GetAllProjects()

	if err != nil {
		return &mevt.MessageEventContent{
			MsgType: mevt.MsgNotice,
			Body:    fmt.Sprint(err),
		}, nil
	}

	var htmlBuffer bytes.Buffer
	var plainBuffer bytes.Buffer
	htmlBuffer.WriteString("<b>Projects:</b><ul>")
	plainBuffer.WriteString("Projects:\n")

	for _, project := range projects {
		htmlBuffer.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s (ID %d)</a></li>", project.URL.Board, project.Name, project.ID))
		plainBuffer.WriteString(fmt.Sprintf("- %s (ID %d, %s)\n", project.Name, project.ID, project.URL.Board))
	}

	htmlBuffer.WriteString("</ul>")

	return &mevt.MessageEventContent{
		MsgType:       mevt.MsgNotice,
		Body:          plainBuffer.String(),
		Format:        mevt.FormatHTML,
		FormattedBody: htmlBuffer.String(),
	}, nil
}

func (k *Service) Commands(cli types.MatrixClient) []types.Command {
	return []types.Command{
		{
			Path: []string{"kanboard", "create"},
			Command: func(roomID id.RoomID, userID id.UserID, args []string) (interface{}, error) {
				return k.cmdKanboardCreateTask(roomID, userID, args)
			},
		},
		{
			Path: []string{"kanboard", "projects"},
			Command: func(roomID id.RoomID, userID id.UserID, args []string) (interface{}, error) {
				return k.cmdKanboardListProjects(roomID, userID, args)
			},
		},
	}
}

func init() {
	types.RegisterService(func(serviceID string, serviceUserID id.UserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
