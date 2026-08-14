package js

type Controller struct {
	Name string
	Path string
}

var TagFilterController = Controller{
	Name: "tag-filter",
	Path: "/assets/js/tag_filter_controller.js",
}

var ControllerPaths = []string{
	TagFilterController.Path,
}
