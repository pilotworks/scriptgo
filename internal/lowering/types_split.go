package lowering

import (
	"strings"
)

func splitTopLevel(s string) []string {
	var parts []string
	var cur strings.Builder
	depthBrace := 0
	depthBracket := 0
	depthAngle := 0
	depthParen := 0

	for _, r := range s {
		switch r {
		case '{':
			depthBrace++
			cur.WriteRune(r)
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
			cur.WriteRune(r)
		case '[':
			depthBracket++
			cur.WriteRune(r)
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
			cur.WriteRune(r)
		case '<':
			depthAngle++
			cur.WriteRune(r)
		case '>':
			if depthAngle > 0 {
				depthAngle--
			}
			cur.WriteRune(r)
		case '(':
			depthParen++
			cur.WriteRune(r)
		case ')':
			if depthParen > 0 {
				depthParen--
			}
			cur.WriteRune(r)
		case ';', ',':
			if depthBrace == 0 && depthBracket == 0 && depthAngle == 0 && depthParen == 0 {
				if cur.Len() > 0 {
					parts = append(parts, strings.TrimSpace(cur.String()))
					cur.Reset()
				}
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		trimmed := strings.TrimSpace(cur.String())
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitTopLevelUnion(s string) []string {
	var parts []string
	var cur strings.Builder
	depthBrace := 0
	depthBracket := 0
	depthAngle := 0
	depthParen := 0

	for _, r := range s {
		switch r {
		case '{':
			depthBrace++
			cur.WriteRune(r)
		case '}':
			depthBrace--
			cur.WriteRune(r)
		case '[':
			depthBracket++
			cur.WriteRune(r)
		case ']':
			depthBracket--
			cur.WriteRune(r)
		case '<':
			depthAngle++
			cur.WriteRune(r)
		case '>':
			depthAngle--
			cur.WriteRune(r)
		case '(':
			depthParen++
			cur.WriteRune(r)
		case ')':
			depthParen--
			cur.WriteRune(r)
		case '|':
			if depthBrace == 0 && depthBracket == 0 && depthAngle == 0 && depthParen == 0 {
				if cur.Len() > 0 {
					trimmed := strings.TrimSpace(cur.String())
					if trimmed != "" {
						parts = append(parts, trimmed)
					}
					cur.Reset()
				}
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		trimmed := strings.TrimSpace(cur.String())
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitTopLevelIntersection(s string) []string {
	var parts []string
	var cur strings.Builder
	depthBrace := 0
	depthBracket := 0
	depthAngle := 0
	depthParen := 0

	for _, r := range s {
		switch r {
		case '{':
			depthBrace++
			cur.WriteRune(r)
		case '}':
			depthBrace--
			cur.WriteRune(r)
		case '[':
			depthBracket++
			cur.WriteRune(r)
		case ']':
			depthBracket--
			cur.WriteRune(r)
		case '<':
			depthAngle++
			cur.WriteRune(r)
		case '>':
			depthAngle--
			cur.WriteRune(r)
		case '(':
			depthParen++
			cur.WriteRune(r)
		case ')':
			depthParen--
			cur.WriteRune(r)
		case '&':
			if depthBrace == 0 && depthBracket == 0 && depthAngle == 0 && depthParen == 0 {
				if cur.Len() > 0 {
					trimmed := strings.TrimSpace(cur.String())
					if trimmed != "" {
						parts = append(parts, trimmed)
					}
					cur.Reset()
				}
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		trimmed := strings.TrimSpace(cur.String())
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
