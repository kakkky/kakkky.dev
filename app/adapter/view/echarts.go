package view

import (
	"html/template"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"

	"github.com/kakkky/kakkky.dev/domain"
)

// RenderDailyLineChart は 日次 UU/PV 2 系列の折れ線を <div>+<script> の snippet として返す。
// echarts.min.js が同一ページ内で読み込まれている前提。
func RenderDailyLineChart(points []domain.AnalyticsDailyPoint) template.HTML {
	line := charts.NewLine()

	dates := make([]string, len(points))
	pv := make([]opts.LineData, len(points))
	uu := make([]opts.LineData, len(points))
	for i, p := range points {
		dates[i] = p.Date.Format("01/02")
		pv[i] = opts.LineData{Value: p.PageViews}
		uu[i] = opts.LineData{Value: p.Users}
	}

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "200px",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Top: "0"}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", Data: dates}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value"}),
		charts.WithGridOpts(opts.Grid{Left: "35", Right: "15", Top: "30", Bottom: "25"}),
	)
	line.SetXAxis(dates).
		AddSeries("PageViews", pv).
		AddSeries("Users", uu).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	return echartsSnippet(line)
}

// RenderSparkline は 記事別テーブルの各行に埋め込む極小折れ線 (親コンテナ幅 x 48px, 軸なし)。
// 幅は "100%" にしているため <td class="w-56"> のような固定幅コンテナに直接収まる。
// 端点のマーカー円が clip されないよう grid の Left/Right に余白を持たせている。
func RenderSparkline(points []domain.AnalyticsDailyPoint) template.HTML {
	line := charts.NewLine()

	dates := make([]string, len(points))
	pv := make([]opts.LineData, len(points))
	uu := make([]opts.LineData, len(points))
	for i, p := range points {
		dates[i] = p.Date.Format("01/02")
		pv[i] = opts.LineData{Value: p.PageViews}
		uu[i] = opts.LineData{Value: p.Users}
	}

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "64px"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithXAxisOpts(opts.XAxis{Show: opts.Bool(false), Type: "category", Data: dates, BoundaryGap: opts.Bool(false)}),
		charts.WithYAxisOpts(opts.YAxis{Show: opts.Bool(false), Type: "value"}),
		charts.WithGridOpts(opts.Grid{Left: "8", Right: "8", Top: "6", Bottom: "6"}),
	)
	line.SetXAxis(dates).
		AddSeries("PV", pv).
		AddSeries("UU", uu)

	// テーブルセル内のチャートは行/列の overflow で tooltip がクリップされがち。
	// appendToBody:true で tooltip DOM を <body> 直下に付け替えて回避する。
	line.AddJSFuncs(`%MY_ECHARTS%.setOption({tooltip: {appendToBody: true}})`)

	return echartsSnippet(line)
}

// echartsSnippet は full-page HTML ではなく <div>+<script> の 2 要素だけを返す。
// go-echarts の Element は `<div class="container">` で始まるが、この class 名は Tailwind の
// container コンポーネント (max-width + margin-auto) と衝突するため strip する。
func echartsSnippet(r render.Renderer) template.HTML {
	s := r.RenderSnippet()
	element := strings.Replace(s.Element, `<div class="container">`, `<div>`, 1)
	element = strings.Replace(element, `<div class="item"`, `<div`, 1)
	return template.HTML(element + s.Script)
}
