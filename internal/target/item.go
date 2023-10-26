package target

import (
	"github.com/AndreZiviani/boundary-fuzzy/internal/run"
	"github.com/hashicorp/boundary/api/targets"
)

type Target struct {
	title       string
	description string
	target      *targets.Target
	session     *SessionInfo
	task        *run.Task
}

func (t Target) Title() string       { return t.title }
func (t Target) Description() string { return t.description }
func (t Target) FilterValue() string { return t.title }
