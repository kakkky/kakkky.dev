package middleware

import (
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
	"github.com/kakkky/kakkky.dev/sentry"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := errors.NewContext(r.Context())
		ctx = sentry.NewContext(ctx, r)
		r = r.WithContext(ctx)

		defer func() {
			var panicked bool
			if p := recover(); p != nil {
				sentry.Recover(ctx, p)
				errors.Set(ctx, domain.ErrInternal)
				panicked = true
			}

			err := errors.FromContext(ctx)
			if err == nil {
				return
			}
			status, msg := errorStatusAndMessage(err)
			if status >= 500 && !panicked {
				sentry.Notify(ctx, err)
			}
			renderError(w, r, err, status, msg)
		}()

		next.ServeHTTP(w, r)
	})
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

// NOTE: エラー画面の Header は PublicBaseURL 空で描画されるため, admin サブドメイン上で
// フルページでエラー画面を返す場合の Feed リンクは "/feed" のままになり、正しく遷移できない問題がある。
func renderError(w http.ResponseWriter, r *http.Request, err error, status int, msg string) {
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

// Turbo は 200/422 の turbo-stream レスポンスのみ消化する。
// 400/409 は 422 に寄せて frame/stream 経由でも banner を反映させる。
func turboStreamStatus(status int) int {
	switch status {
	case http.StatusBadRequest, http.StatusConflict:
		return http.StatusUnprocessableEntity
	}
	return status
}
