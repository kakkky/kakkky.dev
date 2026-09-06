package handler

import (
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostArticleHandler struct {
	updateArticleUsecase *usecase.UpdateArticleUsecase
}

func NewPostArticleHandler(updateArticleUsecase *usecase.UpdateArticleUsecase) *PostArticleHandler {
	return &PostArticleHandler{
		updateArticleUsecase: updateArticleUsecase,
	}
}

func (h *PostArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
		return
	}

	slug := domain.Slug(r.PathValue("slug"))
	tagIDStrs := r.Form["tag_id"]
	newTagNames := r.Form["new_tag"]

	existingTagIDs := make([]domain.TagID, 0, len(tagIDStrs))
	for _, s := range tagIDStrs {
		if s == "" {
			continue
		}
		existingTagIDs = append(existingTagIDs, domain.TagID(s))
	}

	if _, err := h.updateArticleUsecase.Exec(ctx, usecase.UpdateArticleUsecaseInput{
		Slug:           slug,
		Title:          r.FormValue("title"),
		Body:           r.FormValue("body"),
		Status:         domain.ArticleStatus(r.FormValue("status")),
		ExistingTagIDs: existingTagIDs,
		NewTagNames:    newTagNames,
	}); err != nil {
		RenderError(rw, r, err)
		return
	}

	turbo.StreamHeader(rw)
	_ = partials.Flash(partials.FlashViewModel{
		Kind: partials.FlashKindInfo,
		Msg:  "更新しました",
	}).Render(ctx, rw)
}
