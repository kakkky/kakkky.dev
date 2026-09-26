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
// BatchRunReports / RunReport のみ差し替え、他のメソッドは Unimplemented にフォールバック。
type mockAnalyticsServer struct {
	pb.UnimplementedBetaAnalyticsDataServer

	batchMock   func() (*pb.BatchRunReportsResponse, error)
	gotBatchReq *pb.BatchRunReportsRequest

	runMock   func() (*pb.RunReportResponse, error)
	gotRunReq *pb.RunReportRequest
}

func (m *mockAnalyticsServer) BatchRunReports(_ context.Context, req *pb.BatchRunReportsRequest) (*pb.BatchRunReportsResponse, error) {
	m.gotBatchReq = req
	return m.batchMock()
}

func (m *mockAnalyticsServer) RunReport(_ context.Context, req *pb.RunReportRequest) (*pb.RunReportResponse, error) {
	m.gotRunReq = req
	return m.runMock()
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
	t.Parallel()
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
			t.Parallel()
			mock := &mockAnalyticsServer{batchMock: tt.mock}
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
			require.NotNil(t, mock.gotBatchReq)
			assert.Equal(t, "properties/1234567", mock.gotBatchReq.Property)
			require.Len(t, mock.gotBatchReq.Requests, 2)

			// req[0]: date x (activeUsers, screenPageViews) + date asc
			req0 := mock.gotBatchReq.Requests[0]
			require.Len(t, req0.Dimensions, 1)
			assert.Equal(t, "date", req0.Dimensions[0].Name)
			require.Len(t, req0.Metrics, 2)
			assert.Equal(t, "activeUsers", req0.Metrics[0].Name)
			assert.Equal(t, "screenPageViews", req0.Metrics[1].Name)
			require.Len(t, req0.OrderBys, 1)
			assert.False(t, req0.OrderBys[0].Desc)
			assert.Equal(t, "date", req0.OrderBys[0].GetDimension().DimensionName)

			// req[1]: source x medium, sessions desc, Limit 10
			req1 := mock.gotBatchReq.Requests[1]
			require.Len(t, req1.Dimensions, 2)
			assert.Equal(t, "sessionSource", req1.Dimensions[0].Name)
			assert.Equal(t, "sessionMedium", req1.Dimensions[1].Name)
			require.Len(t, req1.Metrics, 1)
			assert.Equal(t, "sessions", req1.Metrics[0].Name)
			assert.Equal(t, int64(10), req1.Limit)
			require.Len(t, req1.OrderBys, 1)
			assert.True(t, req1.OrderBys[0].Desc)
			assert.Equal(t, "sessions", req1.OrderBys[0].GetMetric().MetricName)

			// DateRanges は全 request で同一
			for _, rq := range mock.gotBatchReq.Requests {
				require.Len(t, rq.DateRanges, 1)
				assert.Equal(t, "2026-09-01", rq.DateRanges[0].StartDate)
				assert.Equal(t, "2026-09-30", rq.DateRanges[0].EndDate)
			}
		})
	}
}

func TestGoogleAnalyticsClient_FetchArticleMetrics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		mock        func() (*pb.RunReportResponse, error)
		wantErr     bool
		wantMetrics []domain.AnalyticsArticleMetrics
	}{
		{
			name: "success: pagePathFilter is set and pagePath dimension is appended",
			mock: func() (*pb.RunReportResponse, error) {
				return &pb.RunReportResponse{
					Rows: []*pb.Row{{
						DimensionValues: []*pb.DimensionValue{
							{OneValue: &pb.DimensionValue_Value{Value: "20260901"}},
							{OneValue: &pb.DimensionValue_Value{Value: "/articles/foo"}},
						},
						MetricValues: []*pb.MetricValue{
							{OneValue: &pb.MetricValue_Value{Value: "2"}},
							{OneValue: &pb.MetricValue_Value{Value: "5"}},
						},
					}},
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
				},
			},
		},
		{
			name:    "error: gRPC error is wrapped as domain.ErrInternal",
			mock:    func() (*pb.RunReportResponse, error) { return nil, errors.New("boom") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockAnalyticsServer{runMock: tt.mock}
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

			// request 構造の検証
			require.NotNil(t, mock.gotRunReq)
			assert.Equal(t, "properties/1234567", mock.gotRunReq.Property)

			// Dimensions: date, pagePath
			require.Len(t, mock.gotRunReq.Dimensions, 2)
			assert.Equal(t, "date", mock.gotRunReq.Dimensions[0].Name)
			assert.Equal(t, "pagePath", mock.gotRunReq.Dimensions[1].Name)

			// Metrics: activeUsers, screenPageViews
			require.Len(t, mock.gotRunReq.Metrics, 2)
			assert.Equal(t, "activeUsers", mock.gotRunReq.Metrics[0].Name)
			assert.Equal(t, "screenPageViews", mock.gotRunReq.Metrics[1].Name)

			// pagePath BEGINS_WITH "/articles/" filter
			require.NotNil(t, mock.gotRunReq.DimensionFilter)
			f := mock.gotRunReq.DimensionFilter.GetFilter()
			require.NotNil(t, f)
			assert.Equal(t, "pagePath", f.FieldName)
			sf := f.GetStringFilter()
			require.NotNil(t, sf)
			assert.Equal(t, pb.Filter_StringFilter_BEGINS_WITH, sf.MatchType)
			assert.Equal(t, "/articles/", sf.Value)

			// DateRange
			require.Len(t, mock.gotRunReq.DateRanges, 1)
			assert.Equal(t, "2026-09-01", mock.gotRunReq.DateRanges[0].StartDate)
			assert.Equal(t, "2026-09-30", mock.gotRunReq.DateRanges[0].EndDate)
		})
	}
}
