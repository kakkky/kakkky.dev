package view

import (
	"html/template"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

type LineChartOpts struct {
	Height      string            // chart 高さ (CSS 値。例: "200px")
	XAxisLabels []string          // x 軸のカテゴリラベル (例: 日付文字列の並び)
	Series      []LineChartSeries // 描画する系列 (複数系列を重ねられる)
	ShowLegend  bool              // chart 上部に「色 + 系列名」の対応表 (凡例) を表示
	ShowXAxis   bool              // x 軸 (目盛/ラベル) の表示
	ShowYAxis   bool              // y 軸 (目盛/ラベル) の表示
	Grid        LineChartGrid     // chart 領域内側の padding (px)
}

type LineChartSeries struct {
	Name string  // 凡例/tooltip に出す系列名
	Data []int64 // XAxisLabels と同じ長さの値列
}

type LineChartGrid struct {
	Left   string // grid 左側 padding (y 軸ラベル用に確保)
	Right  string // grid 右側 padding (右端ラベルはみ出し防止)
	Top    string // grid 上側 padding (凡例/タイトル用に確保)
	Bottom string // grid 下側 padding (x 軸ラベル用に確保)
}

// RenderLineChart は echarts の折れ線を <div>+<script> の snippet として返す汎用関数。
func RenderLineChart(o LineChartOpts) template.HTML {
	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: o.Height}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(o.ShowLegend), Top: "0"}),
		charts.WithXAxisOpts(opts.XAxis{
			Show:        opts.Bool(o.ShowXAxis),
			Type:        "category",
			Data:        o.XAxisLabels,
			BoundaryGap: opts.Bool(false),
		}),
		charts.WithYAxisOpts(opts.YAxis{Show: opts.Bool(o.ShowYAxis), Type: "value"}),
		charts.WithGridOpts(opts.Grid{Left: o.Grid.Left, Right: o.Grid.Right, Top: o.Grid.Top, Bottom: o.Grid.Bottom}),
	)

	line.SetXAxis(o.XAxisLabels)
	for _, s := range o.Series {
		data := make([]opts.LineData, len(s.Data))
		for i, v := range s.Data {
			data[i] = opts.LineData{Value: v}
		}
		line.AddSeries(s.Name, data)
	}

	return patchEChartsSnippet(line)
}

// patchEChartsSnippet は go-echarts が吐く <div>+<script> の class 剥がしと tooltip option 注入を行って返す。
func patchEChartsSnippet(r render.Renderer) template.HTML {
	// s.Element = chart 用の <div>、s.Script = 初期化 <script>
	s := r.RenderSnippet()
	// 外側 <div class="container"> は Tailwind の .container と衝突するので class を剥がす
	element := strings.Replace(s.Element, `<div class="container">`, `<div>`, 1)
	// 内側 <div class="item" ...> も同様に class だけ削除 (id / style は温存)
	element = strings.Replace(element, `<div class="item"`, `<div`, 1)
	// tooltip DOM を <body> 直下に付け替えて overflow クリップを回避する。
	// go-echarts の opts.Tooltip に AppendToBody 相当のフィールドが無いため JSON に直接注入する
	script := strings.Replace(
		s.Script,
		`"tooltip":{"show":true,"trigger":"axis"}`,
		`"tooltip":{"show":true,"trigger":"axis","appendToBody":true}`,
		1,
	)
	return template.HTML(element + script)
}
