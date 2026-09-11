package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostSeriesHandler struct {
	createSeriesUsecase *usecase.CreateSeriesUsecase
}

func NewPostSeriesHandler(createSeriesUsecase *usecase.CreateSeriesUsecase) *PostSeriesHandler {
	return &PostSeriesHandler{
		createSeriesUsecase: createSeriesUsecase,
	}
}

func (h *PostSeriesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	tagIDStrs := r.Form["tag_id"]
	newTagNames := r.Form["new_tag"]

	existingTagIDs := make([]domain.TagID, 0, len(tagIDStrs))
	for _, s := range tagIDStrs {
		if s == "" {
			continue
		}
		existingTagIDs = append(existingTagIDs, domain.TagID(s))
	}

	if _, err := h.createSeriesUsecase.Exec(ctx, usecase.CreateSeriesUsecaseInput{
		Title:          title,
		Description:    description,
		ExistingTagIDs: existingTagIDs,
		NewTagNames:    newTagNames,
	}); err != nil {
		RenderError(rw, r, err)
		return
	}

	http.Redirect(rw, r, "/admin/dashboard", http.StatusSeeOther)
}
