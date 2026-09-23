package client

import (
	"context"
	"encoding/base64"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/transport/grpc"
	pb "google.golang.org/genproto/googleapis/analytics/data/v1beta"

	"github.com/kakkky/kakkky.dev/domain"
)

// ref: https://github.com/googleapis/googleapis/blob/master/google/analytics/data/v1beta/analytics_data_api.proto
const (
	gaEndpoint = "analyticsdata.googleapis.com:443"
	gaScope    = "https://www.googleapis.com/auth/analytics.readonly"
)

type GoogleAnalyticsClient struct {
	svc        pb.BetaAnalyticsDataClient
	propertyID string
}

func (c *Client) NewGoogleAnalyticsClient() domain.GoogleAnalyticsClient {
	key, err := base64.StdEncoding.DecodeString(c.cfg.GAServiceAccountKeyB64)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(
		context.Background(),
		option.WithEndpoint(gaEndpoint),
		option.WithAuthCredentialsJSON(option.ServiceAccount, key),
		option.WithScopes(gaScope),
	)
	if err != nil {
		panic(err)
	}
	return &GoogleAnalyticsClient{
		svc: pb.NewBetaAnalyticsDataClient(conn),
		// ref: https://github.com/googleapis/googleapis/blob/master/google/analytics/data/v1beta/analytics_data_api.proto#L617
		propertyID: "properties/" + c.cfg.GAPropertyID,
	}
}

func (a *GoogleAnalyticsClient) FetchSiteMetrics(ctx context.Context, r domain.AnalyticsDateRange) (domain.AnalyticsSiteMetrics, error) {
	req := &pb.BatchRunReportsRequest{
		Property: a.propertyID,
		Requests: []*pb.RunReportRequest{
			// 一定期間内の 日時の UU/PV
			{
				Property:   a.propertyID,
				DateRanges: []*pb.DateRange{toDateRangePb(r)},
				Dimensions: []*pb.Dimension{{Name: "date"}},
				Metrics: []*pb.Metric{
					{Name: "activeUsers"},
					{Name: "screenPageViews"},
				},
				OrderBys: []*pb.OrderBy{{
					OneOrderBy: &pb.OrderBy_Dimension{
						Dimension: &pb.OrderBy_DimensionOrderBy{DimensionName: "date"},
					},
				}},
			},
			// 一定期間内の　Referrerのセッション数 Top10
			{
				Property:   a.propertyID,
				DateRanges: []*pb.DateRange{toDateRangePb(r)},
				Dimensions: []*pb.Dimension{
					{Name: "sessionSource"},
					{Name: "sessionMedium"},
				},
				Metrics: []*pb.Metric{{Name: "sessions"}},
				OrderBys: []*pb.OrderBy{{
					Desc: true,
					OneOrderBy: &pb.OrderBy_Metric{
						Metric: &pb.OrderBy_MetricOrderBy{MetricName: "sessions"},
					},
				}},
				Limit: 10,
			},
		},
	}
	resp, err := a.svc.BatchRunReports(ctx, req)
	if err != nil {
		return domain.AnalyticsSiteMetrics{}, domain.ErrInternal.Wrap(err, "GA 取得 に 失敗 しました")
	}
	return parseSiteMetrics(resp), nil
}

func (a *GoogleAnalyticsClient) FetchArticleMetrics(ctx context.Context, r domain.AnalyticsDateRange) ([]domain.AnalyticsArticleMetrics, error) {
	req := &pb.RunReportRequest{
		Property:   a.propertyID,
		DateRanges: []*pb.DateRange{toDateRangePb(r)},
		Dimensions: []*pb.Dimension{{Name: "date"}, {Name: "pagePath"}},
		Metrics:    []*pb.Metric{{Name: "activeUsers"}, {Name: "screenPageViews"}},
		DimensionFilter: &pb.FilterExpression{
			Expr: &pb.FilterExpression_Filter{
				Filter: &pb.Filter{
					FieldName: "pagePath",
					OneFilter: &pb.Filter_StringFilter_{
						StringFilter: &pb.Filter_StringFilter{
							MatchType: pb.Filter_StringFilter_BEGINS_WITH,
							Value:     "/articles/",
						},
					},
				},
			},
		},
	}
	resp, err := a.svc.RunReport(ctx, req)
	if err != nil {
		return nil, domain.ErrInternal.Wrap(err, "GA 取得 に 失敗 しました")
	}
	return parseArticleMetrics(resp), nil
}

func toDateRangePb(r domain.AnalyticsDateRange) *pb.DateRange {
	return &pb.DateRange{
		StartDate: r.From.UTC().Format("2006-01-02"),
		EndDate:   r.To.UTC().Format("2006-01-02"),
	}
}

func parseSiteMetrics(resp *pb.BatchRunReportsResponse) domain.AnalyticsSiteMetrics {
	var m domain.AnalyticsSiteMetrics

	daily := resp.Reports[0]
	m.ByDate = make([]domain.AnalyticsDailyPoint, 0, len(daily.Rows))
	for _, row := range daily.Rows {
		d, err := time.Parse("20060102", row.DimensionValues[0].GetValue())
		if err != nil {
			continue
		}
		m.ByDate = append(m.ByDate, domain.AnalyticsDailyPoint{
			Date:      d,
			Users:     parseInt64(row.MetricValues[0].GetValue()),
			PageViews: parseInt64(row.MetricValues[1].GetValue()),
		})
	}
	for _, p := range m.ByDate {
		m.TotalUsers += p.Users
		m.TotalPageViews += p.PageViews
	}

	referrers := resp.Reports[1]
	m.TopReferrers = make([]domain.AnalyticsReferrerCount, 0, len(referrers.Rows))
	for _, row := range referrers.Rows {
		m.TopReferrers = append(m.TopReferrers, domain.AnalyticsReferrerCount{
			Source:   trimGAParens(row.DimensionValues[0].GetValue()),
			Medium:   trimGAParens(row.DimensionValues[1].GetValue()),
			Sessions: parseInt64(row.MetricValues[0].GetValue()),
		})
	}
	return m
}

func parseArticleMetrics(resp *pb.RunReportResponse) []domain.AnalyticsArticleMetrics {
	bySlug := map[domain.Slug]*domain.AnalyticsArticleMetrics{}

	for _, row := range resp.Rows {
		d, err := time.Parse("20060102", row.DimensionValues[0].GetValue())
		if err != nil {
			continue
		}
		path := row.DimensionValues[1].GetValue()
		slug := domain.Slug(strings.TrimPrefix(path, "/articles/"))
		users := parseInt64(row.MetricValues[0].GetValue())
		pv := parseInt64(row.MetricValues[1].GetValue())

		am, ok := bySlug[slug]
		if !ok {
			am = &domain.AnalyticsArticleMetrics{Slug: slug}
			bySlug[slug] = am
		}
		am.Users += users
		am.PageViews += pv
		am.ByDate = append(am.ByDate, domain.AnalyticsDailyPoint{
			Date:      d,
			Users:     users,
			PageViews: pv,
		})
	}

	out := make([]domain.AnalyticsArticleMetrics, 0, len(bySlug))
	for _, am := range bySlug {
		out = append(out, *am)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PageViews != out[j].PageViews {
			return out[i].PageViews > out[j].PageViews
		}
		return out[i].Slug < out[j].Slug
	})
	return out
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// trimGAParens は GA4 が Source / Medium の特別値を "(direct)" / "(none)" /
// "(not set)" のように括弧付きで返すため、括弧を除去する
func trimGAParens(s string) string {
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		return s[1 : len(s)-1]
	}
	return s
}
