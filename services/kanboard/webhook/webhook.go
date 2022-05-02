package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	mevt "maunium.net/go/mautrix/event"
)

type RequestData struct {
	EventName   string          `json:"event_name"`
	EventData   json.RawMessage `json:"event_data"`
	EventAuthor string          `json:"event_author"`
}

type TaskMoveColumnEvent struct {
	TaskId      int                   `json:"task_id"`
	Task        Task                  `json:"task"`
	Changes     TaskMoveColumnChanges `json:"changes"`
	ProjectId   int                   `json:"project_id"`
	Position    int                   `json:"position"`
	ColumnId    int                   `json:"column_id,string"`
	SwimlaneId  int                   `json:"swimlane_id,string"`
	SrcColumnId int                   `json:"src_column_id"`
	DstColumnId int                   `json:"dst_column_id,string"` // API inconsistency
	DateMoved   int                   `json:"date_moved"`
}

type TaskMoveColumnChanges struct {
	SrcColumnId int `json:"src_column_id"`
	DstColumnId int `json:"dst_column_id,string"` // API inconsistency
	DateMoved   int `json:"date_moved"`
}

type TaskAssigneeChangeEvent struct {
	TaskId  int                       `json:"task_id"`
	Task    Task                      `json:"task"`
	Changes TaskAssigneeChangeChanges `json:"changes"`
}

type TaskAssigneeChangeChanges struct {
	OwnerId int `json:"owner_id,string"`
}

type Task struct {
	Id               int    `json:"id"`
	Reference        string `json:"reference"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	DateCreation     int    `json:"date_creation"`
	ProjectId        int    `json:"project_id"`
	DateDue          int    `json:"date_due"`
	SwimlaneName     string `json:"swimlane_name"`
	ProjectName      string `json:"project_name"`
	ColumnTitle      string `json:"column_title"`
	AssigneeUsername string `json:"assignee_username"`
	AssigneeName     string `json:"assignee_name"`
	CreatorUsername  string `json:"creator_username"`
	CreatorName      string `json:"creator_name"`
}

// return eventName, projectID, message
func OnReceiveRequest(r *http.Request, secretToken string, endpoint string) (string, int, *mevt.MessageEventContent, error) {
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read request body: %s", err)
		return "", 0, nil, err
	}

	var req RequestData
	if err := json.Unmarshal(bodyData, &req); err != nil {
		return "", 0, nil, err
	}

	switch req.EventName {
	case "task.move.column":
		var tmc TaskMoveColumnEvent
		if err := json.Unmarshal([]byte(req.EventData), &tmc); err != nil {
			return "", 0, nil, err
		}

		return req.EventName, tmc.Task.ProjectId, &mevt.MessageEventContent{
			MsgType:       mevt.MsgNotice,
			Body:          fmt.Sprintf("Task %s moved to %s by %s", tmc.Task.Title, tmc.Task.ColumnTitle, req.EventAuthor),
			Format:        mevt.FormatHTML,
			FormattedBody: fmt.Sprintf("Task <a href=\"%s/task/%d\">%s</a> moved to <b>%s</b> <i>by %s</i>", endpoint, tmc.Task.Id, tmc.Task.Title, tmc.Task.ColumnTitle, req.EventAuthor),
		}, nil

	case "task.assignee_change":
		var tac TaskAssigneeChangeEvent
		if err := json.Unmarshal([]byte(req.EventData), &tac); err != nil {
			return "", 0, nil, err
		}

		return req.EventName, tac.Task.ProjectId, &mevt.MessageEventContent{
			MsgType:       mevt.MsgNotice,
			Body:          fmt.Sprintf("Task %s assigned to %s by %s", tac.Task.Title, tac.Task.AssigneeName, req.EventAuthor),
			Format:        mevt.FormatHTML,
			FormattedBody: fmt.Sprintf("Task <a href=\"%s/task/%d\">%s</a> assigned to <b>%s</b> <i>by %s</i>", endpoint, tac.Task.Id, tac.Task.Title, tac.Task.AssigneeName, req.EventAuthor),
		}, nil
	}

	log.Errorf("Unknown event name: %s", req.EventName)
	return "", 0, nil, nil
}
