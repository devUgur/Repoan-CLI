package output

import (
	"encoding/json"
	"strings"

	"github.com/repoan/repoan/internal/model"
)

type TreeOptions struct {
	Format string
	ASCII  bool
}

func FormatTree(node *model.FileItem, opts TreeOptions) string {
	switch strings.ToLower(opts.Format) {
	case "md":
		return "```\n" + renderText(node, "", true, true, opts.ASCII) + "```"
	case "json":
		data, err := json.MarshalIndent(node, "", "  ")
		if err != nil {
			return "error marshaling to json"
		}
		return string(data)
	default:
		return renderText(node, "", true, true, opts.ASCII)
	}
}

func renderText(node *model.FileItem, prefix string, isLast bool, isRoot bool, useASCII bool) string {
	var sb strings.Builder

	if isRoot {
		sb.WriteString(node.Name + "\n")
		for i, child := range node.Children {
			isChildLast := i == len(node.Children)-1
			sb.WriteString(renderText(child, "", isChildLast, false, useASCII))
		}
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if useASCII {
			connector = "|-- "
			if isLast {
				connector = "`-- "
			}
		}

		sb.WriteString(prefix + connector + node.Name + "\n")

		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			if useASCII {
				newPrefix += "|   "
			} else {
				newPrefix += "│   "
			}
		}

		for i, child := range node.Children {
			isChildLast := i == len(node.Children)-1
			sb.WriteString(renderText(child, newPrefix, isChildLast, false, useASCII))
		}
	}

	return sb.String()
}
