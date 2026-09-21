package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "google.golang.org/genproto/googleapis/analytics/data/v1beta"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
)

// mockAnalyticsServer は in-memory gRPC server 用の pb.BetaAnalyticsDataServer 実装。
// BatchRunReports のみ差し替え、他のメソッドは Unimplemented にフォールバック。
type mockAnalyticsServer struct {
	pb.UnimplementedBetaAnalyticsDataServer
	mock   func() (*pb.BatchRunReportsResponse, error)
	gotReq *pb.BatchRunReportsRequest
}

func (m *mockAnalyticsServer) BatchRunReports(_ context.Context, req *pb.BatchRunReportsRequest) (*pb.BatchRunReportsResponse, error) {
	m.gotReq = req
	return m.mock()
}

func newMockGAClient(t *testing.T, propertyID string, mock *mockAnalyticsServer) *GoogleAnalyticsClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	pb.RegisterBetaAnalyticsDataServer(srv, mock)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop()
	})

	return &GoogleAnalyticsClient{
		svc:        pb.NewBetaAnalyticsDataClient(conn),
		propertyID: propertyID,
	}
}

func TestGoogleAnalyticsClient_FetchSiteMetrics(t *testing.T) {
	tests := []struct {
		name        string
		mock        func() (*pb.BatchRunReportsResponse, error)
		wantErr     bool
		wantMetrics domain.AnalyticsSiteMetrics
	}{
		{
			name: "success: request is constructed correctly and response is parsed",
			mock: func() (*pb.BatchRunReportsResponse, error) {
				return &pb.BatchRunReportsResponse{
					Reports: []*pb.RunReportResponse{
						{Rows: []*pb.Row{{
							DimensionValues: []*pb.DimensionValue{
								{OneValue: &pb.DimensionValue_Value{Value: "20260901"}},
							},
							MetricValues: []*pb.MetricValue{
								{OneValue: &pb.MetricValue_Value{Value: "3"}},
								{OneValue: &pb.MetricValue_Value{Value: "9"}},
							},
						}}},
						{Rows: []*pb.Row{{
							DimensionValues: []*pb.DimensionValue{
								{OneValue: &pb.DimensionValue_Value{Value: "0"}},
								{OneValue: &pb.DimensionValue_Value{Value: "3"}},
							},
							MetricValues: []*pb.MetricValue{
								{OneValue: &pb.MetricValue_Value{Value: "1"}},
								{OneValue: &pb.MetricValue_Value{Value: "2"}},
							},
						}}},
						{Rows: []*pb.Row{{
							DimensionValues: []*pb.DimensionValue{
								{OneValue: &pb.DimensionValue_Value{Value: "google"}},
								{OneValue: &pb.DimensionValue_Value{Value: "organic"}},
							},
							MetricValues: []*pb.MetricValue{
								{OneValue: &pb.MetricValue_Value{Value: "42"}},
							},
						}}},
					},
				}, nil
			},
			wantMetrics: domain.AnalyticsSiteMetrics{
				TotalUsers:     3,
				TotalPageViews: 9,
				ByDate: []domain.AnalyticsDailyPoint{
					{Date: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Users: 3, PageViews: 9},
				},
				ByHour: []domain.AnalyticsHourlyPoint{
					{Weekday: time.Sunday, Hour: 3, Users: 1, PageViews: 2},
				},
				TopReferrers: []domain.AnalyticsReferrerCount{
					{Source: "google", Medium: "organic", Sessions: 42},
				},
			},
		},
		{
			name:    "error: gRPC error is wrapped as domain.ErrInternal",
			mock:    func() (*pb.BatchRunReportsResponse, error) { return nil, errors.New("boom") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAnalyticsServer{mock: tt.mock}
			c := newMockGAClient(t, "properties/1234567", mock)

			r := domain.AnalyticsDateRange{
				From: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
			}
			got, err := c.FetchSiteMetrics(context.Background(), r)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, domain.ErrInternal))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMetrics, got)

			// request 構造の検証
			require.NotNil(t, mock.gotReq)
			assert.Equal(t, "properties/1234567", mock.gotReq.Property)
			require.Len(t, mock.gotReq.Requests, 3)

			// req[0]: date x (activeUsers, screenPageViews) + date asc
			req0 := mock.gotReq.Requests[0]
			require.Len(t, req0.Dimensions, 1)
			assert.Equal(t, "date", req0.Dimensions[0].Name)
			require.Len(t, req0.Metrics, 2)
			assert.Equal(t, "activeUsers", req0.Metrics[0].Name)
			assert.Equal(t, "screenPageViews", req0.Metrics[1].Name)
			require.Len(t, req0.OrderBys, 1)
			assert.False(t, req0.OrderBys[0].Desc)
			assert.Equal(t, "date", req0.OrderBys[0].GetDimension().DimensionName)

			// req[1]: dayOfWeek x hour
			req1 := mock.gotReq.Requests[1]
			require.Len(t, req1.Dimensions, 2)
			assert.Equal(t, "dayOfWeek", req1.Dimensions[0].Name)
			assert.Equal(t, "hour", req1.Dimensions[1].Name)

			// req[2]: source x medium, sessions desc, Limit 10
			req2 := mock.gotReq.Requests[2]
			require.Len(t, req2.Dimensions, 2)
			assert.Equal(t, "sessionSource", req2.Dimensions[0].Name)
			assert.Equal(t, "sessionMedium", req2.Dimensions[1].Name)
			require.Len(t, req2.Metrics, 1)
			assert.Equal(t, "sessions", req2.Metrics[0].Name)
			assert.Equal(t, int64(10), req2.Limit)
			require.Len(t, req2.OrderBys, 1)
			assert.True(t, req2.OrderBys[0].Desc)
			assert.Equal(t, "sessions", req2.OrderBys[0].GetMetric().MetricName)

			// DateRanges は全 request で同一
			for _, rq := range mock.gotReq.Requests {
				require.Len(t, rq.DateRanges, 1)
				assert.Equal(t, "2026-09-01", rq.DateRanges[0].StartDate)
				assert.Equal(t, "2026-09-30", rq.DateRanges[0].EndDate)
			}
		})
	}
}

func TestGoogleAnalyticsClient_FetchArticleMetrics(t *testing.T) {
	tests := []struct {
		name        string
		mock        func() (*pb.BatchRunReportsResponse, error)
		wantErr     bool
		wantMetrics []domain.AnalyticsArticleMetrics
	}{
		{
			name: "success: pagePathFilter is set on all requests and pagePath dimension is appended",
			mock: func() (*pb.BatchRunReportsResponse, error) {
				return &pb.BatchRunReportsResponse{
					Reports: []*pb.RunReportResponse{
						{Rows: []*pb.Row{{
							DimensionValues: []*pb.DimensionValue{
								{OneValue: &pb.DimensionValue_Value{Value: "20260901"}},
								{OneValue: &pb.DimensionValue_Value{Value: "/articles/foo"}},
							},
							MetricValues: []*pb.MetricValue{
								{OneValue: &pb.MetricValue_Value{Value: "2"}},
								{OneValue: &pb.MetricValue_Value{Value: "5"}},
							},
						}}},
						{Rows: []*pb.Row{{
							DimensionValues: []*pb.DimensionValue{
								{OneValue: &pb.DimensionValue_Value{Value: "0"}},
								{OneValue: &pb.DimensionValue_Value{Value: "9"}},
								{OneValue: &pb.DimensionValue_Value{Value: "/articles/foo"}},
							},
							MetricValues: []*pb.MetricValue{
								{OneValue: &pb.MetricValue_Value{Value: "1"}},
								{OneValue: &pb.MetricValue_Value{Value: "2"}},
							},
						}}},
						{Rows: []*pb.Row{{
							DimensionValues: []*pb.DimensionValue{
								{OneValue: &pb.DimensionValue_Value{Value: "google"}},
								{OneValue: &pb.DimensionValue_Value{Value: "organic"}},
								{OneValue: &pb.DimensionValue_Value{Value: "/articles/foo"}},
							},
							MetricValues: []*pb.MetricValue{
								{OneValue: &pb.MetricValue_Value{Value: "3"}},
							},
						}}},
					},
				}, nil
			},
			wantMetrics: []domain.AnalyticsArticleMetrics{
				{
					Slug:      "foo",
					Users:     2,
					PageViews: 5,
					ByDate: []domain.AnalyticsDailyPoint{
						{Date: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Users: 2, PageViews: 5},
					},
					ByHour: []domain.AnalyticsHourlyPoint{
						{Weekday: time.Sunday, Hour: 9, Users: 1, PageViews: 2},
					},
					TopReferrers: []domain.AnalyticsReferrerCount{
						{Source: "google", Medium: "organic", Sessions: 3},
					},
				},
			},
		},
		{
			name:    "error: gRPC error is wrapped as domain.ErrInternal",
			mock:    func() (*pb.BatchRunReportsResponse, error) { return nil, errors.New("boom") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAnalyticsServer{mock: tt.mock}
			c := newMockGAClient(t, "properties/1234567", mock)

			r := domain.AnalyticsDateRange{
				From: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
			}
			got, err := c.FetchArticleMetrics(context.Background(), r)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, domain.ErrInternal))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMetrics, got)

			// 全 request に pagePath BEGINS_WITH "/articles/" filter が付いている
			require.Len(t, mock.gotReq.Requests, 3)
			for _, rq := range mock.gotReq.Requests {
				require.NotNil(t, rq.DimensionFilter)
				f := rq.DimensionFilter.GetFilter()
				require.NotNil(t, f)
				assert.Equal(t, "pagePath", f.FieldName)
				sf := f.GetStringFilter()
				require.NotNil(t, sf)
				assert.Equal(t, pb.Filter_StringFilter_BEGINS_WITH, sf.MatchType)
				assert.Equal(t, "/articles/", sf.Value)
			}

			// 各 request の Dimensions 末尾に pagePath が含まれる
			req0 := mock.gotReq.Requests[0]
			require.Len(t, req0.Dimensions, 2)
			assert.Equal(t, "date", req0.Dimensions[0].Name)
			assert.Equal(t, "pagePath", req0.Dimensions[1].Name)

			req1 := mock.gotReq.Requests[1]
			require.Len(t, req1.Dimensions, 3)
			assert.Equal(t, "dayOfWeek", req1.Dimensions[0].Name)
			assert.Equal(t, "hour", req1.Dimensions[1].Name)
			assert.Equal(t, "pagePath", req1.Dimensions[2].Name)

			req2 := mock.gotReq.Requests[2]
			require.Len(t, req2.Dimensions, 3)
			assert.Equal(t, "sessionSource", req2.Dimensions[0].Name)
			assert.Equal(t, "sessionMedium", req2.Dimensions[1].Name)
			assert.Equal(t, "pagePath", req2.Dimensions[2].Name)
		})
	}
}
