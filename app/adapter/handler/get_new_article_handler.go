package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetNewArticleHandler struct {
	listTagsUsecase *usecase.ListTagsUsecase
}

func NewGetNewArticleHandler(listTagsUsecase *usecase.ListTagsUsecase) *GetNewArticleHandler {
	return &GetNewArticleHandler{
		listTagsUsecase: listTagsUsecase,
	}
}

func (h *GetNewArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	out, err := h.listTagsUsecase.Exec(ctx)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	vm := pages.NewArticleViewModel{
		Form: partials.ArticleFormViewModel{
			Action:       "/admin/articles",
			SubmitLabel:  "作成",
			Status:       string(domain.ArticleStatusDraft),
			AllTagNames:  tagNames(out.Tags),
			SelectedTags: []string{},
		},
	}
	_ = pages.NewArticle(vm).Render(ctx, rw)
}

func tagNames(tags []domain.Tag) []string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}
