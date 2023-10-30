package target

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/AndreZiviani/boundary-fuzzy/internal/run"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/sessions"
	"github.com/hashicorp/boundary/api/targets"
)

type SessionInfo struct {
	Address         string                       `json:"address"`
	Port            int                          `json:"port"`
	Protocol        string                       `json:"protocol"`
	Expiration      time.Time                    `json:"expiration"`
	ConnectionLimit int32                        `json:"connection_limit"`
	SessionId       string                       `json:"session_id"`
	Credentials     []*targets.SessionCredential `json:"credentials,omitempty"`
}

func ConnectToTarget(target *Target) (*run.Task, *exec.Cmd, *SessionInfo, error) {
	task := run.RunTask("boundary", []string{"connect", "-target-id", target.target.Id, "-format", "json"})

	var session SessionInfo

	reader := bufio.NewReader(task.Output)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, nil, fmt.Errorf(msg)
	}

	tmp := strings.NewReader(msg)
	d := json.NewDecoder(tmp)
	d.DisallowUnknownFields()
	err = d.Decode(&session)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(msg)
	}

	var cmd *exec.Cmd
	if port, ok := target.target.Attributes["default_port"]; ok {
		switch int(port.(float64)) {
		case 5432:
			// postgres
			args := []string{
				"-h", "127.0.0.1",
				"-p", strconv.Itoa(session.Port),
				"-d", "postgres",
			}

			if len(session.Credentials) > 0 {
				args = append(args, "-U", session.Credentials[0].Secret.Decoded["username"].(string))
			}

			cmd = exec.Command("psql", args...)
			cmd.Env = os.Environ()

			if len(session.Credentials) > 0 {
				cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", session.Credentials[0].Secret.Decoded["password"]))
			}

		case 3306:
			// mysql
			args := []string{
				"-A",
				"-h", "127.0.0.1",
				"-P", strconv.Itoa(session.Port),
			}

			if len(session.Credentials) > 0 {
				args = append(args, "-u", session.Credentials[0].Secret.Decoded["username"].(string))
			}
			args = append(args, "information_schema")

			cmd = exec.Command("mysql", args...)
			cmd.Env = os.Environ()

			if len(session.Credentials) > 0 {
				cmd.Env = append(cmd.Env, fmt.Sprintf("MYSQL_PWD=%s", session.Credentials[0].Secret.Decoded["password"]))
			}

		case 6379:
			// redis
			cmd = exec.Command(
				"redis-cli",
				"-p", strconv.Itoa(session.Port),
			)
			cmd.Env = os.Environ()

		default:
			// do nothing, just connect to target
		}
	}

	return task, cmd, &session, nil
}

func TerminateSession(boundaryClient *api.Client, session *SessionInfo, task *run.Task) {
	task.Cancel()
	sessionClient := sessions.NewClient(boundaryClient)

	sessionInfo, err := sessionClient.Read(context.TODO(), session.SessionId)
	if err != nil {
		return
	}

	sessionClient.Cancel(context.TODO(), session.SessionId, sessionInfo.Item.Version)

	task.Cmd.Wait()
}
