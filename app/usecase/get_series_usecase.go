package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetSeriesUsecase struct {
	seriesRepo  domain.SeriesRepository
	articleRepo domain.ArticleRepository
	tagRepo     domain.TagRepository
}

func (us *UseCase) NewGetSeriesUsecase() *GetSeriesUsecase {
	return &GetSeriesUsecase{
		seriesRepo:  us.repo.NewSeriesRepository(),
		articleRepo: us.repo.NewArticleRepository(),
		tagRepo:     us.repo.NewTagRepository(),
	}
}

type GetSeriesUsecaseInput struct {
	Slug domain.Slug
}

type GetSeriesUsecaseOutput struct {
	Series   domain.Series
	Articles []GetSeriesUsecaseSeriesArticle
	Tags     map[domain.TagID]domain.Tag
}

type GetSeriesUsecaseSeriesArticle struct {
	Article  domain.Article
	Position int
}

func (us *GetSeriesUsecase) Exec(ctx context.Context, in GetSeriesUsecaseInput) (GetSeriesUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return GetSeriesUsecaseOutput{}, err
	}
	return getSeries(ctx, us.seriesRepo, us.articleRepo, us.tagRepo, in.Slug, false)
}

func (in GetSeriesUsecaseInput) validate() error {
	if in.Slug == "" {
		return domain.ErrInvalidArgument.With("slug は 必須 です")
	}
	return nil
}

// getSeries は series 取得 の 共通ロジック。
// includeDraft=false の 場合、draft series は ErrNotFound として 扱い、
// article 一覧 も published の みに 絞る (公開 面 の 挙動)。
func getSeries(
	ctx context.Context,
	seriesRepo domain.SeriesRepository,
	articleRepo domain.ArticleRepository,
	tagRepo domain.TagRepository,
	slug domain.Slug,
	includeDraft bool,
) (GetSeriesUsecaseOutput, error) {
	series, err := seriesRepo.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return GetSeriesUsecaseOutput{}, domain.ErrNotFound.With("series が 見つかりません")
		}
		return GetSeriesUsecaseOutput{}, err
	}
	if series.Status == domain.SeriesStatusDraft && !includeDraft {
		return GetSeriesUsecaseOutput{}, domain.ErrNotFound.With("series が 見つかりません")
	}

	articleIDs := make([]domain.ArticleID, 0, len(series.Articles))
	positionByID := make(map[domain.ArticleID]int, len(series.Articles))
	for _, sa := range series.Articles {
		articleIDs = append(articleIDs, sa.ArticleID)
		positionByID[sa.ArticleID] = sa.Position
	}

	articles, err := articleRepo.FindByIDs(ctx, articleIDs...)
	if err != nil {
		return GetSeriesUsecaseOutput{}, err
	}

	seriesArticles := make([]GetSeriesUsecaseSeriesArticle, 0, len(articles))
	for _, a := range articles {
		if !includeDraft && a.Status != domain.ArticleStatusPublished {
			continue
		}
		seriesArticles = append(seriesArticles, GetSeriesUsecaseSeriesArticle{
			Article:  *a,
			Position: positionByID[a.ID],
		})
	}
	slices.SortFunc(seriesArticles, func(a, b GetSeriesUsecaseSeriesArticle) int {
		return a.Position - b.Position
	})

	tagIDSet := map[domain.TagID]struct{}{}
	for _, tid := range series.TagIDs {
		tagIDSet[tid] = struct{}{}
	}
	for _, sa := range seriesArticles {
		for _, tid := range sa.Article.TagIDs {
			tagIDSet[tid] = struct{}{}
		}
	}
	tagIDs := make([]domain.TagID, 0, len(tagIDSet))
	for tid := range tagIDSet {
		tagIDs = append(tagIDs, tid)
	}
	tags, err := tagRepo.FindByIDs(ctx, tagIDs...)
	if err != nil {
		return GetSeriesUsecaseOutput{}, err
	}
	tagsByID := make(map[domain.TagID]domain.Tag, len(tags))
	for _, t := range tags {
		tagsByID[t.ID] = *t
	}

	return GetSeriesUsecaseOutput{
		Series:   *series,
		Articles: seriesArticles,
		Tags:     tagsByID,
	}, nil
}
