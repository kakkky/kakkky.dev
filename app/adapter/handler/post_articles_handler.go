package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostArticlesHandler struct {
	createArticleUsecase *usecase.CreateArticleUsecase
}

func NewPostArticlesHandler(createArticleUsecase *usecase.CreateArticleUsecase) *PostArticlesHandler {
	return &PostArticlesHandler{
		createArticleUsecase: createArticleUsecase,
	}
}

func (h *PostArticlesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
		return
	}

	title := r.FormValue("title")
	tagIDStrs := r.Form["tag_id"]
	newTagNames := r.Form["new_tag"]

	existingTagIDs := make([]domain.TagID, 0, len(tagIDStrs))
	for _, s := range tagIDStrs {
		if s == "" {
			continue
		}
		existingTagIDs = append(existingTagIDs, domain.TagID(s))
	}

	out, err := h.createArticleUsecase.Exec(ctx, usecase.CreateArticleUsecaseInput{
		Title:          title,
		ExistingTagIDs: existingTagIDs,
		NewTagNames:    newTagNames,
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	http.Redirect(rw, r, "/admin/articles/"+string(out.ArticleSlug)+"/edit", http.StatusSeeOther)
}
