package handler

import (
	"errors"
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
)

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := errorStatusAndMessage(err)
	isTurbo := turbo.IsFrameRequest(r) || turbo.IsStreamRequest(r)

	if isTurbo && !errors.Is(err, domain.ErrInternal) {
		turbo.StreamHeader(w)
		w.WriteHeader(turboStreamStatus(status))
		_ = partials.Flash(partials.FlashViewModel{
			Kind: partials.FlashKindErr,
			Msg:  msg,
		}).Render(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = pages.Error(pages.ErrorViewModel{
		Status: status,
		Msg:    msg,
	}).Render(r.Context(), w)
}

func errorStatusAndMessage(err error) (int, string) {
	var domainErr *domain.Error
	var domainErrMsg string
	if errors.As(err, &domainErr) && domainErr.Message() != "" {
		domainErrMsg = domainErr.Message()
	}

	var httpStatus int
	var fallbackErrMsg string

	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		httpStatus = http.StatusBadRequest
		fallbackErrMsg = "入力に誤りがあります"
	case errors.Is(err, domain.ErrAlreadyExists):
		httpStatus = http.StatusConflict
		fallbackErrMsg = "既に 存在 します"
	case errors.Is(err, domain.ErrNotFound):
		httpStatus = http.StatusNotFound
		fallbackErrMsg = "ページが見つかりません"
	default:
		httpStatus = http.StatusInternalServerError
		fallbackErrMsg = "サーバーエラーが発生しました。時間をおいてお試しください。"

		// ドメインエラーメッセージはUIでも表示する可能性がある。機密情報の漏れを防ぐため空にしておく
		domainErrMsg = ""
	}

	if domainErrMsg != "" {
		return httpStatus, domainErrMsg
	}
	return httpStatus, fallbackErrMsg
}

// Turbo は 200/422 の turbo-stream レスポンスのみ消化する。
// 400/409 は 422 に寄せて frame/stream 経由でも banner を反映させる。
func turboStreamStatus(status int) int {
	switch status {
	case http.StatusBadRequest, http.StatusConflict:
		return http.StatusUnprocessableEntity
	}
	return status
}
