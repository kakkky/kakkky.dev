package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

const uniqueSlugMaxAttempts = 100

type CreateArticleUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewCreateArticleUsecase() *CreateArticleUsecase {
	return &CreateArticleUsecase{
		repo: us.repo,
	}
}

type CreateArticleUsecaseInput struct {
	Title    string
	Body     string
	Status   string
	TagNames []string
}

type CreateArticleUsecaseOutput struct {
	Slug domain.Slug
}

func (us *CreateArticleUsecase) Exec(ctx context.Context, in CreateArticleUsecaseInput) (CreateArticleUsecaseOutput, error) {
	status := domain.ArticleStatus(in.Status)

	var publishedAt time.Time
	if status == domain.ArticleStatusPublished {
		publishedAt = time.Now().UTC()
	}

	// slug は title から server 側で導出。placeholder として base を渡して
	// NewArticle の validation を通し、tx 内で unique な slug に差し替える
	base := domain.DeriveSlug(in.Title)
	article, err := domain.NewArticle(base, in.Title, in.Body, status, publishedAt)
	if err != nil {
		return CreateArticleUsecaseOutput{}, err
	}

	err = us.repo.WithTx(ctx, func(tx domain.Repository) error {
		unique, err := uniqueArticleSlug(ctx, tx.NewArticleRepository(), base)
		if err != nil {
			return err
		}
		article.Slug = unique

		tagIDs, err := resolveTagIDs(ctx, tx.NewTagRepository(), in.TagNames)
		if err != nil {
			return err
		}
		if err := article.AddTags(tagIDs); err != nil {
			return err
		}
		return tx.NewArticleRepository().Create(ctx, article)
	})
	if err != nil {
		return CreateArticleUsecaseOutput{}, err
	}

	return CreateArticleUsecaseOutput{Slug: article.Slug}, nil
}

// uniqueArticleSlug は base 候補から始めて、既存 slug と衝突しない slug を返す。
// 衝突時は base-2, base-3, ... と suffix を付ける。base が長い場合は末尾を trim する。
func uniqueArticleSlug(ctx context.Context, ar domain.ArticleRepository, base domain.Slug) (domain.Slug, error) {
	candidate := base
	for attempt := 1; attempt <= uniqueSlugMaxAttempts; attempt++ {
		_, err := ar.FindBySlug(ctx, candidate)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return candidate, nil
			}
			return "", err
		}
		// 衝突: base-{attempt+1}
		suffix := "-" + strconv.Itoa(attempt+1)
		maxBase := domain.SlugMaxLength - len(suffix)
		b := string(base)
		if len(b) > maxBase {
			b = b[:maxBase]
			b = strings.TrimRight(b, "-")
		}
		candidate = domain.Slug(b + suffix)
	}
	return "", domain.ErrInternal.With("ユニークな slug を生成できませんでした")
}

// resolveTagIDs は name 一覧を tagID に解決する。
// 既存 tag があれば流用し、無ければその場で作成する。
// 入力順を保ちつつ、重複 name は除去する。
func resolveTagIDs(ctx context.Context, tagRepo domain.TagRepository, names []string) ([]domain.TagID, error) {
	uniqNames := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		uniqNames = append(uniqNames, n)
	}
	if len(uniqNames) == 0 {
		return nil, nil
	}

	existing, err := tagRepo.FindByNames(ctx, uniqNames...)
	if err != nil {
		return nil, err
	}
	byName := make(map[string]*domain.Tag, len(existing))
	for _, t := range existing {
		byName[t.Name] = t
	}

	tagIDs := make([]domain.TagID, 0, len(uniqNames))
	for _, n := range uniqNames {
		if t, ok := byName[n]; ok {
			tagIDs = append(tagIDs, t.ID)
			continue
		}
		slug, err := domain.NewTagSlug(n)
		if err != nil {
			return nil, err
		}
		tag, err := domain.NewTag(slug, n)
		if err != nil {
			return nil, err
		}
		if err := tagRepo.Create(ctx, tag); err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, tag.ID)
	}
	return tagIDs, nil
}
