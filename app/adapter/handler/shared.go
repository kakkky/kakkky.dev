package handler

import (
	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/domain"
)

func toTagViewModels(tags []domain.Tag) []components.TagViewModel {
	out := make([]components.TagViewModel, len(tags))
	for i, t := range tags {
		out[i] = components.TagViewModel{ID: string(t.ID), Slug: string(t.Slug), Name: t.Name}
	}
	return out
}
