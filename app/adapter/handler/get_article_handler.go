package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetArticleHandler struct {
	getArticleUsecase *usecase.GetArticleUsecase
}

func NewGetArticleHandler(getArticleUsecase *usecase.GetArticleUsecase) *GetArticleHandler {
	return &GetArticleHandler{
		getArticleUsecase: getArticleUsecase,
	}
}

func (h *GetArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	out, err := h.getArticleUsecase.Exec(ctx, usecase.GetArticleUsecaseInput{Slug: slug})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	tagNames := make([]string, 0, len(out.Article.TagIDs))
	for _, tid := range out.Article.TagIDs {
		if t, ok := out.Tags[tid]; ok {
			tagNames = append(tagNames, t.Name)
		}
	}
	_ = tagNames

	// vm := pages.ArticleViewModel{
	// 	Title:       out.Article.Title,
	// 	PublishedAt: out.Article.PublishedAt,
	// 	Tags:        tagNames,
	// 	BodyMD:      out.Article.Body,
	// }
	// _ = pages.Article(vm).Render(ctx, rw)
}
