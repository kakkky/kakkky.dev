package handler

import (
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetLinkPreviewHandler struct {
	getLinkPreviewUsecase *usecase.GetLinkPreviewUsecase
}

func NewGetLinkPreviewHandler(getLinkPreviewUsecase *usecase.GetLinkPreviewUsecase) *GetLinkPreviewHandler {
	return &GetLinkPreviewHandler{
		getLinkPreviewUsecase: getLinkPreviewUsecase,
	}
}

func (h *GetLinkPreviewHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rawURL := r.URL.Query().Get("url")

	frameID := turbo.FrameID(r)
	if frameID == "" {
		// 直接アクセス (curl / browser で /link-preview?url=... 直打ち) 時の fallback
		frameID = "link-preview"
	}

	vm := components.LinkPreviewCardViewModel{URL: rawURL}

	// fetch エラーは握りつぶし、空 vm で card を描画する
	if out, err := h.getLinkPreviewUsecase.Exec(ctx, usecase.GetLinkPreviewUsecaseInput{URL: rawURL}); err == nil {
		vm.Host = out.Data.Host
		vm.Title = out.Data.Title
		vm.Description = out.Data.Description
		vm.Image = out.Data.Image
	}

	_ = partials.LinkPreviewFrame(frameID, vm).Render(ctx, rw)
}
