package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
	"github.com/kakkky/kakkky.dev/usecase"
)

const postImagesMaxMemory = 1 << 20

type PostImagesHandler struct {
	uploadImageUsecase *usecase.UploadImageUsecase
}

func NewPostImagesHandler(uploadImageUsecase *usecase.UploadImageUsecase) *PostImagesHandler {
	return &PostImagesHandler{
		uploadImageUsecase: uploadImageUsecase,
	}
}

type postImagesResponse struct {
	URL string `json:"url"`
}

func (h *PostImagesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(postImagesMaxMemory); err != nil {
		errors.Set(r.Context(), domain.ErrInvalidArgument.Wrap(err, "画像 の フォーム を 解釈 できません"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		errors.Set(r.Context(), domain.ErrInvalidArgument.Wrap(err, "画像 ファイル が 見つかりません"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	out, err := h.uploadImageUsecase.Exec(ctx, usecase.UploadImageUsecaseInput{
		ContentType: contentType,
		Size:        header.Size,
		Body:        file,
	})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}

	body, err := json.Marshal(postImagesResponse{URL: out.URL})
	if err != nil {
		errors.Set(r.Context(), domain.ErrInternal.Wrap(err, "レスポンス の 生成 に 失敗 しました"))
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = rw.Write(body)
}
