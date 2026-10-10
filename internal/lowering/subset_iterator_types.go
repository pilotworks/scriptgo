package lowering

import "strings"

// iteratorTypeNames are the lib iterator interfaces whose TReturn and TNext
// type parameters default to any (`Generator<T = unknown, TReturn = any,
// TNext = any>`), so the checker prints `Generator<string, any, any>` for a
// declared `Generator<string>`.
var iteratorTypeNames = []string{"AsyncIterableIterator", "AsyncIteratorObject", "IterableIterator", "AsyncGenerator", "IteratorObject", "AsyncIterator", "Generator", "Iterator"}

// withoutIteratorDefaultAny drops trailing any type arguments after the
// first from iterator interface types: they are the lib defaults for the
// return and next types, not an any value the program holds. An any that
// reaches a value (IteratorResult<T, any>.value) is still reported there.
func withoutIteratorDefaultAny(typ string) string {
	var out strings.Builder
	for i := 0; i < len(typ); {
		name := iteratorTypeNameAt(typ, i)
		if name == "" {
			out.WriteByte(typ[i])
			i++
			continue
		}
		open := i + len(name)
		end := matchingAngle(typ, open)
		if end < 0 {
			out.WriteString(typ[i:])
			break
		}
		args := splitTopLevel(typ[open+1 : end])
		for len(args) > 1 && strings.TrimSpace(args[len(args)-1]) == "any" {
			args = args[:len(args)-1]
		}
		for index := range args {
			args[index] = withoutIteratorDefaultAny(strings.TrimSpace(args[index]))
		}
		out.WriteString(name + "<" + strings.Join(args, ", ") + ">")
		i = end + 1
	}
	return out.String()
}

// iteratorTypeNameAt is the iterator interface name that starts a generic
// type reference at typ[i], or "".
func iteratorTypeNameAt(typ string, i int) string {
	if i > 0 && isTypeNameByte(typ[i-1]) {
		return ""
	}
	for _, name := range iteratorTypeNames {
		if strings.HasPrefix(typ[i:], name+"<") {
			return name
		}
	}
	return ""
}

// matchingAngle is the index of the '>' closing the '<' at typ[open], or -1.
func matchingAngle(typ string, open int) int {
	depth := 0
	for i := open; i < len(typ); i++ {
		switch typ[i] {
		case '<':
			depth++
		case '>':
			if i > 0 && typ[i-1] == '=' {
				continue // the arrow of a function type
			}
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func isTypeNameByte(b byte) bool {
	return b == '_' || b == '$' || b == '.' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
