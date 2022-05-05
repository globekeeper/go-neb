package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	log "github.com/sirupsen/logrus"
	mevt "maunium.net/go/mautrix/event"
)

type RequestData struct {
	EventName   string          `json:"event_name"`
	EventData   json.RawMessage `json:"event_data"`
	EventAuthor string          `json:"event_author"`
}

type CommentCreateEvent struct {
	Comment Comment `json:"comment"`
	Task    Task    `json:"task"`
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

type TaskCreateEvent struct {
	TaskId int  `json:"task_id"`
	Task   Task `json:"task"`
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

type Comment struct {
	Id               int    `json:"id"`
	TaskId           int    `json:"task_id"`
	UserId           int    `json:"user_id"`
	DateCreation     int    `json:"date_creation"`
	DateModification int    `json:"date_modification"`
	Comment          string `json:"comment"`
	Reference        string `json:"reference"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	AvatarPath       string `json:"avatar_path"`
}

func taskUrl(endpoint string, taskId int) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("tasks/%d", taskId))
	return u.String(), nil
}

// return eventName, projectID, message, error
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

	case "comment.create":
		var evtData CommentCreateEvent
		if err := json.Unmarshal([]byte(req.EventData), &evtData); err != nil {
			return "", 0, nil, err
		}

		taskUrl, err := taskUrl(endpoint, evtData.Task.Id)
		if err != nil {
			return "", 0, nil, err
		}

		return req.EventName, evtData.Task.ProjectId, &mevt.MessageEventContent{
			MsgType:       mevt.MsgNotice,
			Body:          fmt.Sprintf("%s commented on task %s: %s", evtData.Comment.Username, evtData.Task.Title, evtData.Comment.Comment),
			Format:        mevt.FormatHTML,
			FormattedBody: fmt.Sprintf("<i>%s commented on task</i> <a href=\"%s\">%s</a>: %s", evtData.Comment.Username, taskUrl, evtData.Task.Title, evtData.Comment.Comment),
		}, nil

	case "task.move.column":
		var evtData TaskMoveColumnEvent
		if err := json.Unmarshal([]byte(req.EventData), &evtData); err != nil {
			return "", 0, nil, err
		}

		taskUrl, err := taskUrl(endpoint, evtData.Task.Id)
		if err != nil {
			return "", 0, nil, err
		}

		return req.EventName, evtData.Task.ProjectId, &mevt.MessageEventContent{
			MsgType:       mevt.MsgNotice,
			Body:          fmt.Sprintf("Task %s moved to %s by %s", evtData.Task.Title, evtData.Task.ColumnTitle, req.EventAuthor),
			Format:        mevt.FormatHTML,
			FormattedBody: fmt.Sprintf("Task <a href=\"%s\">%s</a> moved to <b>%s</b> <i>by %s</i>", taskUrl, evtData.Task.Title, evtData.Task.ColumnTitle, req.EventAuthor),
		}, nil

	case "task.create":
		var evtData TaskCreateEvent
		if err := json.Unmarshal([]byte(req.EventData), &evtData); err != nil {
			return "", 0, nil, err
		}
		taskUrl, err := taskUrl(endpoint, evtData.Task.Id)
		if err != nil {
			return "", 0, nil, err
		}

		return req.EventName, evtData.Task.ProjectId, &mevt.MessageEventContent{
			MsgType:       mevt.MsgNotice,
			Body:          fmt.Sprintf("Task %s created by %s", evtData.Task.Title, req.EventAuthor),
			Format:        mevt.FormatHTML,
			FormattedBody: fmt.Sprintf("Task <a href=\"%s\">%s</a> created <i>by %s</i>", taskUrl, evtData.Task.Title, req.EventAuthor),
		}, nil

	case "task.assignee_change":
		var evtData TaskAssigneeChangeEvent
		if err := json.Unmarshal([]byte(req.EventData), &evtData); err != nil {
			return "", 0, nil, err
		}
		taskUrl, err := taskUrl(endpoint, evtData.Task.Id)
		if err != nil {
			return "", 0, nil, err
		}

		return req.EventName, evtData.Task.ProjectId, &mevt.MessageEventContent{
			MsgType:       mevt.MsgNotice,
			Body:          fmt.Sprintf("Task %s assigned to %s by %s", evtData.Task.Title, evtData.Task.AssigneeName, req.EventAuthor),
			Format:        mevt.FormatHTML,
			FormattedBody: fmt.Sprintf("Task <a href=\"%s\">%s</a> assigned to <b>%s</b> <i>by %s</i>", taskUrl, evtData.Task.Title, evtData.Task.AssigneeName, req.EventAuthor),
		}, nil
	}

	log.Errorf("Unknown event name: %s", req.EventName)
	return "", 0, nil, nil
}
