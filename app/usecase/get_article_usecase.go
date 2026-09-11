package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetArticleUsecase struct {
	articleRepo domain.ArticleRepository
	seriesRepo  domain.SeriesRepository
	tagRepo     domain.TagRepository
}

func (us *UseCase) NewGetArticleUsecase() *GetArticleUsecase {
	return &GetArticleUsecase{
		articleRepo: us.repo.NewArticleRepository(),
		seriesRepo:  us.repo.NewSeriesRepository(),
		tagRepo:     us.repo.NewTagRepository(),
	}
}

type GetArticleUsecaseInput struct {
	Slug domain.Slug
}

type GetArticleUsecaseOutput struct {
	Article     domain.Article
	Tags        map[domain.TagID]domain.Tag
	InSeriesRef *GetArticleUsecaseSeriesRef
}

type GetArticleUsecaseSeriesRef struct {
	Slug                   domain.Slug
	Title                  string
	Status                 domain.SeriesStatus
	CurrentArticlePosition int
	PrevArticleRef         *GetArticleUsecaseSeriesNeighborArticleRef
	NextArticleRef         *GetArticleUsecaseSeriesNeighborArticleRef
}

type GetArticleUsecaseSeriesNeighborArticleRef struct {
	Slug     domain.Slug
	Title    string
	Position int
}

func (us *GetArticleUsecase) Exec(ctx context.Context, input GetArticleUsecaseInput) (GetArticleUsecaseOutput, error) {
	if err := input.validate(); err != nil {
		return GetArticleUsecaseOutput{}, err
	}
	article, err := us.articleRepo.FindBySlug(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return GetArticleUsecaseOutput{}, domain.ErrNotFound.With("article が 見つかりません")
		}
		return GetArticleUsecaseOutput{}, err
	}
	tags, err := us.tagRepo.FindByIDs(ctx, article.TagIDs...)
	if err != nil {
		return GetArticleUsecaseOutput{}, err
	}
	tagsByID := make(map[domain.TagID]domain.Tag, len(tags))
	for _, tag := range tags {
		tagsByID[tag.ID] = *tag
	}

	inSeriesRef, err := us.resolveSeriesRef(ctx, article.ID)
	if err != nil {
		return GetArticleUsecaseOutput{}, err
	}

	return GetArticleUsecaseOutput{
		Article:     *article,
		Tags:        tagsByID,
		InSeriesRef: inSeriesRef,
	}, nil
}

func (us *GetArticleUsecase) resolveSeriesRef(ctx context.Context, articleID domain.ArticleID) (*GetArticleUsecaseSeriesRef, error) {
	series, err := us.seriesRepo.FindByArticleID(ctx, articleID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if series.Status == domain.SeriesStatusDraft {
		return nil, nil
	}

	// series.Articles は position 昇順に整列済み
	currentIdx := slices.IndexFunc(series.Articles, func(sa domain.SeriesArticle) bool {
		return sa.ArticleID == articleID
	})
	if currentIdx < 0 {
		return nil, domain.ErrInternal.With("series に含まれるはずの article が 見つかりません")
	}

	var prevSA, nextSA *domain.SeriesArticle
	if currentIdx > 0 {
		prevSA = &series.Articles[currentIdx-1]
	}
	if currentIdx+1 < len(series.Articles) {
		nextSA = &series.Articles[currentIdx+1]
	}

	neighborIDs := make([]domain.ArticleID, 0, 2)
	if prevSA != nil {
		neighborIDs = append(neighborIDs, prevSA.ArticleID)
	}
	if nextSA != nil {
		neighborIDs = append(neighborIDs, nextSA.ArticleID)
	}
	neighborArticles, err := us.articleRepo.FindByIDs(ctx, neighborIDs...)
	if err != nil {
		return nil, err
	}
	neighborArticleByID := make(map[domain.ArticleID]*domain.Article, len(neighborArticles))
	for _, a := range neighborArticles {
		neighborArticleByID[a.ID] = a
	}

	var prev, next *GetArticleUsecaseSeriesNeighborArticleRef
	if prevSA != nil {
		if a, ok := neighborArticleByID[prevSA.ArticleID]; ok && a.Status == domain.ArticleStatusPublished {
			prev = &GetArticleUsecaseSeriesNeighborArticleRef{Slug: a.Slug, Title: a.Title, Position: prevSA.Position}
		}
	}
	if nextSA != nil {
		if a, ok := neighborArticleByID[nextSA.ArticleID]; ok && a.Status == domain.ArticleStatusPublished {
			next = &GetArticleUsecaseSeriesNeighborArticleRef{Slug: a.Slug, Title: a.Title, Position: nextSA.Position}
		}
	}

	return &GetArticleUsecaseSeriesRef{
		Slug:                   series.Slug,
		Title:                  series.Title,
		Status:                 series.Status,
		CurrentArticlePosition: series.Articles[currentIdx].Position,
		PrevArticleRef:         prev,
		NextArticleRef:         next,
	}, nil
}

func (in GetArticleUsecaseInput) validate() error {
	if in.Slug == "" {
		return domain.ErrInvalidArgument.With("slug は 必須 です")
	}
	return nil
}
