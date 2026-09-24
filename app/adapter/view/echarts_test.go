package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderLineChart(t *testing.T) {
	tests := []struct {
		name        string
		opts        LineChartOpts
		wantContain []string
	}{
		{
			name: "success: full featured site line chart",
			opts: LineChartOpts{
				Height:      "200px",
				XAxisLabels: []string{"09/01", "09/02"},
				Series: []LineChartSeries{
					{Name: "PageViews", Data: []int64{9, 12}},
					{Name: "Users", Data: []int64{3, 4}},
				},
				ShowLegend: true,
				ShowXAxis:  true,
				ShowYAxis:  true,
				Grid:       LineChartGrid{Left: "35", Right: "15", Top: "30", Bottom: "25"},
			},
			wantContain: []string{
				`width:100%`,
				`height:200px`,
				`echarts.init`,
				`PageViews`,
				`Users`,
				`"appendToBody":true`,
			},
		},
		{
			name: "success: trend line style (bare)",
			opts: LineChartOpts{
				Height:      "64px",
				XAxisLabels: []string{"09/01"},
				Series:      []LineChartSeries{{Name: "PV", Data: []int64{1}}},
				Grid:        LineChartGrid{Left: "8", Right: "8", Top: "6", Bottom: "6"},
			},
			wantContain: []string{
				`width:100%`,
				`height:64px`,
				`"boundaryGap":false`,
				`"appendToBody":true`,
			},
		},
		{
			name: "success: empty series and labels return non-empty snippet without panicking",
			opts: LineChartOpts{Height: "40px"},
			wantContain: []string{
				`echarts.init`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(RenderLineChart(tt.opts))
			require.NotEmpty(t, got)
			// Tailwind の container / item コンポーネント class と衝突しないよう strip されている
			assert.NotContains(t, got, `class="container"`)
			assert.NotContains(t, got, `class="item"`)
			for _, want := range tt.wantContain {
				assert.Contains(t, got, want)
			}
		})
	}
}

