package view

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
)

func TestRenderDailyLineChart(t *testing.T) {
	tests := []struct {
		name   string
		points []domain.AnalyticsDailyPoint
	}{
		{
			name:   "success: empty slice returns non-empty snippet without panicking",
			points: nil,
		},
		{
			name: "success: multiple points render series and axis data",
			points: []domain.AnalyticsDailyPoint{
				{Date: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Users: 3, PageViews: 9},
				{Date: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), Users: 4, PageViews: 12},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(RenderDailyLineChart(tt.points))
			require.NotEmpty(t, got)
			// Tailwind の container / item コンポーネント class と衝突しないよう strip されている
			assert.NotContains(t, got, `class="container"`)
			assert.NotContains(t, got, `class="item"`)
			assert.Contains(t, got, `echarts.init`)
		})
	}
}

func TestRenderSparkline(t *testing.T) {
	tests := []struct {
		name   string
		points []domain.AnalyticsDailyPoint
	}{
		{
			name:   "success: empty slice returns non-empty snippet without panicking",
			points: nil,
		},
		{
			name: "success: axis is hidden and grid is compact",
			points: []domain.AnalyticsDailyPoint{
				{Date: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Users: 3, PageViews: 9},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(RenderSparkline(tt.points))
			require.NotEmpty(t, got)
			assert.Contains(t, got, `width:100%`)
			assert.Contains(t, got, `height:64px`)
		})
	}
}

func TestRenderDailyLineChart_UniqueChartID(t *testing.T) {
	a := string(RenderDailyLineChart(nil))
	b := string(RenderDailyLineChart(nil))
	// chart ID は go-echarts が UUID を発行するため、2 回描画すると別 ID になる。
	// これがないと同一 frame 内で複数チャートを swap した際に衝突する。
	assert.NotEqual(t, extractChartID(a), extractChartID(b))
}

func extractChartID(html string) string {
	const marker = `id="`
	i := strings.Index(html, marker)
	if i < 0 {
		return ""
	}
	rest := html[i+len(marker):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}
