package output

import (
	"encoding/json"
	"strings"

	"github.com/repoan/repoan/internal/model"
)

func FormatTree(node *model.FileItem, format string) string {
	switch strings.ToLower(format) {
	case "md":
		return "```\n" + renderText(node, "", true, true) + "```"
	case "json":
		data, err := json.MarshalIndent(node, "", "  ")
		if err != nil {
			return "error marshaling to json"
		}
		return string(data)
	default:
		return renderText(node, "", true, true)
	}
}

func renderText(node *model.FileItem, prefix string, isLast bool, isRoot bool) string {
	var sb strings.Builder

	if isRoot {
		sb.WriteString(node.Name + "\n")
		for i, child := range node.Children {
			isChildLast := i == len(node.Children)-1
			sb.WriteString(renderText(child, "", isChildLast, false))
		}
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		sb.WriteString(prefix + connector + node.Name + "\n")

		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		for i, child := range node.Children {
			isChildLast := i == len(node.Children)-1
			sb.WriteString(renderText(child, newPrefix, isChildLast, false))
		}
	}

	return sb.String()
}
