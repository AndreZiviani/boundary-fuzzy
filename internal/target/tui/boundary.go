package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/hashicorp/boundary/api/targets"
)

func (t *tui) refreshTargets() error {
	targetsResult, err := t.targetsClient.List(t.ctx, "global", targets.WithRecursive(true))
	if err != nil {
		return err
	}

	tuiTargets := make([]list.Item, 0, len(targetsResult.Items))
	for _, target := range targetsResult.Items {
		tuiTargets = append(
			tuiTargets,
			&Target{
				title:          fmt.Sprintf("%s (%s)", target.Name, target.Scope.Name),
				description:    target.Description,
				target:         target,
				sessionsClient: t.sessionsClient,
				targetClient:   t.targetsClient,
			})
	}

	t.tabs[targetsView].SetItems(tuiTargets)

	return t.refreshFavoriteList()
}
