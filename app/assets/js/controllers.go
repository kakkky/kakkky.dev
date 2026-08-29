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

// TurboStreamActionsPath は Stimulus controller ではなく、Turbo Streams の
// custom action を登録する module script。 layout 側で個別に <script> tag を
// 出す (ScriptLoad の対象外)。
const TurboStreamActionsPath = "/assets/js/turbo_stream_actions.js"

var ArticleFormController = Controller{
	Name: "article-form",
	Path: "/assets/js/article_form_controller.js",
}

var TagComboboxController = Controller{
	Name: "tag-combobox",
	Path: "/assets/js/tag_combobox_controller.js",
}

var ControllerPaths = []string{
	TagFilterController.Path,
	OutlineDialogController.Path,
	ArticleFormController.Path,
	TagComboboxController.Path,
}
