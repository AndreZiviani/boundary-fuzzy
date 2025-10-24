package tui

import (
	"context"
	"fmt"
	"net/netip"
	"os/exec"
	"strconv"
	"time"

	"github.com/AndreZiviani/boundary-fuzzy/internal/run"
	tea "github.com/charmbracelet/bubbletea"
	apiproxy "github.com/hashicorp/boundary/api/proxy"
	"github.com/hashicorp/boundary/api/sessions"
	"github.com/hashicorp/boundary/api/targets"
	"go.uber.org/atomic"
)

type Target struct {
	targetClient   *targets.Client
	sessionsClient *sessions.Client
	title          string
	rtitle         string
	description    string
	rdescription   string
	target         *targets.Target
	session        *SessionInfo
	task           *run.Task
}

func (t Target) Title(tab sessionState) (string, string) {
	switch tab {
	case connectedView:
		return t.title, t.rtitle
	}

	return t.title, ""
}
func (t Target) Description(tab sessionState) (string, string) {
	switch tab {
	case connectedView:
		return t.description, t.rdescription
	}

	return t.description, ""
}

func (t Target) FilterValue() string { return t.title }

type SessionInfo struct {
	ctx                context.Context
	cancel             context.CancelFunc
	clientProxyCloseCh chan struct{}
	active             bool

	sessionClient *sessions.Client

	authorizationToken string
	Address            string
	Port               int
	Expiration         time.Time
	ConnectionLimit    int32
	SessionId          string
	Credentials        []*targets.SessionCredential
}

func (t *Target) Connect() (*exec.Cmd, error) {
	err := t.newSessionProxy(context.Background())
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if port, ok := t.target.Attributes["default_port"]; ok {
		switch int(port.(float64)) {
		case 5432:
			cmd = NewPSQLCommand("127.0.0.1", t.session.Port, t.session.Credentials)

		case 3306:
			cmd = NewMySQLCommand("127.0.0.1", t.session.Port, t.session.Credentials)

		case 6379:
			cmd = NewRedisCommand("127.0.0.1", t.session.Port, t.session.Credentials)

		case 9440:
			cmd = NewClickHouseCommand("127.0.0.1", t.session.Port, t.session.Credentials)

		default:
			// do nothing, just connect to target
		}
	}

	return cmd, nil
}

func (t *Target) newSessionProxy(mainCtx context.Context) error {
	session, err := t.targetClient.AuthorizeSession(mainCtx, t.target.Id)
	if err != nil {
		return err
	}

	auth, err := session.GetSessionAuthorization()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(mainCtx)
	si := &SessionInfo{
		ctx:    ctx,
		cancel: cancel,
		active: true,

		sessionClient:      t.sessionsClient,
		authorizationToken: auth.AuthorizationToken,
		Expiration:         auth.Expiration,
		ConnectionLimit:    auth.ConnectionLimit,
		SessionId:          auth.SessionId,
		Credentials:        auth.Credentials,
	}

	addr, err := netip.ParseAddr("127.0.0.1")
	if err != nil {
		return err
	}

	listenAddr := netip.AddrPortFrom(addr, 0)

	connsLeftCh := make(chan int32)
	apiProxyOpts := []apiproxy.Option{
		apiproxy.WithConnectionsLeftCh(connsLeftCh),
		apiproxy.WithListenAddrPort(listenAddr),
	}

	clientProxy, err := apiproxy.New(
		ctx,
		auth.AuthorizationToken,
		apiProxyOpts...,
	)
	if err != nil {
		return err
	}

	clientProxyCloseCh := make(chan struct{})
	connCountCloseCh := make(chan struct{})

	proxyError := new(atomic.Error)
	go func() {
		defer close(clientProxyCloseCh)
		proxyError.Store(clientProxy.Start())
		si.active = false
	}()
	go func() {
		defer close(connCountCloseCh)
		for {
			select {
			case <-ctx.Done():
				// When the proxy exits it will cancel this even if we haven't
				// done it manually
				return
			case connsLeft := <-connsLeftCh:
				if connsLeft == 0 {
					return
				}
			}
		}
	}()

	listenerCtx, listenerCancel := context.WithTimeout(ctx, 5*time.Second)
	defer listenerCancel()
	proxyAddr := clientProxy.ListenerAddress(listenerCtx)
	if listenerCtx.Err() != nil {
		cancel()
		proxyErr := proxyError.Load()
		if proxyErr != nil {
			return fmt.Errorf("could not start proxy: %w", proxyErr)
		}
		return fmt.Errorf("could not start proxy listener: %w", listenerCtx.Err())
	}
	clientProxyHost, clientProxyPort, err := SplitHostPort(proxyAddr)
	if err != nil {
		return err
	}

	si.Address = clientProxyHost
	si.Port, _ = strconv.Atoi(clientProxyPort)
	si.clientProxyCloseCh = clientProxyCloseCh

	t.session = si

	t.rtitle = fmt.Sprintf("(%s)", clientProxyPort)
	t.rdescription = fmt.Sprintf("(%s)", si.Expiration.Local().Format(time.RFC3339))

	return nil
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
		status := "connected"
		if t.session.ctx.Err() != nil {
			status = t.session.ctx.Err().Error()
		} else if !t.session.active {
			status = "disconnected"
		}

		msg = fmt.Sprintf(
			"%s\n"+
				"Port: %d\n"+
				"Expiration: %s\n"+
				"Session Id: %s\n"+
				"Status: %s\n",
			msg, t.session.Port, t.session.Expiration, t.session.SessionId, status,
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

func (s *SessionInfo) Terminate(task *run.Task) {
	ctx := context.Background()
	s.cancel()

	sessionInfo, err := s.sessionClient.Read(ctx, s.SessionId)
	if err != nil {
		return
	}

	s.sessionClient.Cancel(ctx, s.SessionId, sessionInfo.Item.Version)
}

func (t *Target) Shell(callbackFn tea.ExecCallback) (tea.Cmd, error) {
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
			t.session.Terminate(t.task)
			if err != nil {
				return callbackFn(err)
			}
			return nil
		},
	), nil
}

func (t *Target) IsConnected() bool {
	if t.session == nil {
		return false
	}

	if t.session.ctx.Err() != nil {
		return false
	}

	if !t.session.active {
		return false
	}

	return true
}
