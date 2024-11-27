package tui

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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/boundary/api/sessions"
	"github.com/hashicorp/boundary/api/targets"
)

type Target struct {
	sessionClient *sessions.Client
	title         string
	description   string
	target        *targets.Target
	session       *SessionInfo
	task          *run.Task
}

func (t Target) Title() string       { return t.title }
func (t Target) Description() string { return t.description }
func (t Target) FilterValue() string { return t.title }

type SessionInfo struct {
	sessionClient   *sessions.Client
	Address         string                       `json:"address"`
	Port            int                          `json:"port"`
	Protocol        string                       `json:"protocol"`
	Expiration      time.Time                    `json:"expiration"`
	ConnectionLimit int32                        `json:"connection_limit"`
	SessionId       string                       `json:"session_id"`
	Credentials     []*targets.SessionCredential `json:"credentials,omitempty"`
}

func (t *Target) Connect() (*exec.Cmd, error) {
	task := run.RunTask("boundary", []string{"connect", "-target-id", t.target.Id, "-format", "json"})
	t.task = task

	var session SessionInfo

	reader := bufio.NewReader(task.Output)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf(msg)
	}

	tmp := strings.NewReader(msg)
	d := json.NewDecoder(tmp)
	d.DisallowUnknownFields()
	err = d.Decode(&session)
	if err != nil {
		return nil, fmt.Errorf(msg)
	}

	t.session = &session
	t.session.sessionClient = t.sessionClient

	var cmd *exec.Cmd
	if port, ok := t.target.Attributes["default_port"]; ok {
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

	return cmd, nil
}

func (t Target) Info() string {
	msg := fmt.Sprintf(
		"Scope: %s\n"+
			"Scope Description: %s\n"+
			"Name: %s\n"+
			"Description: %s\n",
		t.target.Scope.Name, t.target.Scope.Description, t.target.Name, t.target.Description,
	)

	if t.session != nil {
		msg = fmt.Sprintf(
			"%s\n"+
				"Port: %d\n"+
				"Expiration: %s\n"+
				"Session Id: %s\n",
			msg, t.session.Port, t.session.Expiration, t.session.SessionId,
		)

		if len(t.session.Credentials) > 0 {
			msg = fmt.Sprintf(
				"%s\n"+
					"Dynamic Credentials:\n"+
					"  Username: %s\n"+
					"  Password: %s\n",
				msg, t.session.Credentials[0].Secret.Decoded["username"], t.session.Credentials[0].Secret.Decoded["password"],
			)
		}
	}

	return msg
}

func (s *SessionInfo) Terminate(ctx context.Context, task *run.Task) {
	task.Cancel()
	sessionInfo, err := s.sessionClient.Read(ctx, s.SessionId)
	if err != nil {
		return
	}

	s.sessionClient.Cancel(ctx, s.SessionId, sessionInfo.Item.Version)

	task.Cmd.Wait()
}

func (t *Target) Shell(ctx context.Context, callbackFn tea.ExecCallback) (tea.Cmd, error) {
	cmd, err := t.Connect()
	if err != nil {
		return nil, err
	}

	if cmd == nil {
		// we are trying to connect to a target that we could not identify its type or does not have a client (e.g. HTTP)
		// just connect to it without opening a shell
		//TODO: show error message
		return nil, nil
	}

	return tea.ExecProcess(
		cmd,
		func(err error) tea.Msg {
			t.session.Terminate(ctx, t.task)
			if err != nil {
				return callbackFn(err)
				// m.previousState = m.state
				// m.state = errorView
				// m.message = err.Error()
			}
			return nil
		},
	), nil
}
