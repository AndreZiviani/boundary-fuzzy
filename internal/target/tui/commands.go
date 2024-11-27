package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/hashicorp/boundary/api/targets"
)

func NewPSQLCommand(host string, port int, credentials []*targets.SessionCredential) *exec.Cmd {
	args := []string{
		"-h", host,
		"-p", strconv.Itoa(port),
		"-d", "postgres",
	}

	if len(credentials) > 0 {
		args = append(args, "-U", credentials[0].Secret.Decoded["username"].(string))
	}

	cmd := exec.Command("psql", args...)
	cmd.Env = os.Environ()

	if len(credentials) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", credentials[0].Secret.Decoded["password"]))
	}

	return cmd
}

func NewMySQLCommand(host string, port int, credentials []*targets.SessionCredential) *exec.Cmd {
	args := []string{
		"-A",
		"-h", host,
		"-P", strconv.Itoa(port),
	}

	if len(credentials) > 0 {
		args = append(args, "-u", credentials[0].Secret.Decoded["username"].(string))
	}
	args = append(args, "information_schema")

	cmd := exec.Command("mysql", args...)
	cmd.Env = os.Environ()

	if len(credentials) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("MYSQL_PWD=%s", credentials[0].Secret.Decoded["password"]))
	}

	return cmd
}

func NewRedisCommand(host string, port int, credentials []*targets.SessionCredential) *exec.Cmd {
	cmd := exec.Command(
		"redis-cli",
		"-p", strconv.Itoa(port),
	)
	cmd.Env = os.Environ()

	return cmd
}
