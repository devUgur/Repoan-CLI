package output

import (
	"github.com/repoan/repoan/internal/scan"
	"strings"
)

func FormatTree(node *scan.FileNode, format string) string {
	switch strings.ToLower(format) {
	case "md":
		return "```\n" + renderText(node, "", true, true) + "```"
	case "json":
		return "// TODO: Implement JSON format"
	default:
		return renderText(node, "", true, true)
	}
}

func renderText(node *scan.FileNode, prefix string, isLast bool, isRoot bool) string {
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
