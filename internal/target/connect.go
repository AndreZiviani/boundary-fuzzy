package target

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

func ConnectToTarget(target Target) (*run.Task, *exec.Cmd, string, error) {
	task := run.RunTask("boundary", []string{"connect", "-target-id", target.target.Id, "-format", "json"})

	task.Output.Scan()

	var session SessionInfo
	err := json.Unmarshal(task.Output.Bytes(), &session)
	if err != nil {
		return nil, nil, "", err
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

	return task, cmd, session.SessionId, nil
}

func TerminateSession(boundaryClient *api.Client, sessionID string, task *run.Task) {
	sessionClient := sessions.NewClient(boundaryClient)

	session, err := sessionClient.Read(context.TODO(), sessionID)
	if err != nil {
		return
	}

	sessionClient.Cancel(context.TODO(), sessionID, session.Item.Version)

	task.Cancel()
	task.Cmd.Wait()
}
