package partials

import "encoding/json"

// tagNamesJSON は tag 名一覧を Stimulus Value (Array 型) 用の JSON にする。
// tagsValue に data-*-tags-value="[...]" として渡す。
func tagNamesJSON(names []string) string {
	if len(names) == 0 {
		return "[]"
	}
	b, err := json.Marshal(names)
	if err != nil {
		return "[]"
	}
	return string(b)
}
