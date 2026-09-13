package js

type Controller struct {
	Name string
	Path string
}

var TagFilterController = Controller{
	Name: "tag-filter",
	Path: "/assets/js/tag_filter_controller.js",
}

var OutlineDialogController = Controller{
	Name: "outline-dialog",
	Path: "/assets/js/outline_dialog_controller.js",
}

var TagInputController = Controller{
	Name: "tag-input",
	Path: "/assets/js/tag_input_controller.js",
}

var ArticleEditorController = Controller{
	Name: "article-editor",
	Path: "/assets/js/article_editor_controller.js",
}

var EditArticleController = Controller{
	Name: "edit-article",
	Path: "/assets/js/edit_article_controller.js",
}

var EditSeriesArticlesController = Controller{
	Name: "edit-series-articles",
	Path: "/assets/js/edit_series_articles_controller.js",
}

