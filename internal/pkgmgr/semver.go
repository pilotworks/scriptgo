package pkgmgr

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Version is the SemVer core used by package resolution. Pre-release and
// build metadata are deliberately rejected until the registry layer needs them.
type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string { return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch) }

func ParseVersion(value string) (Version, error) {
	parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(value, "v")), ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid semantic version %q", value)
	}
	result := Version{}
	fields := []*int{&result.Major, &result.Minor, &result.Patch}
	for i, part := range parts {
		if part == "" || strings.ContainsAny(part, "-+") {
			return Version{}, fmt.Errorf("invalid semantic version %q", value)
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return Version{}, fmt.Errorf("invalid semantic version %q", value)
		}
		*fields[i] = number
	}
	return result, nil
}

func CompareVersions(left, right Version) int {
	if left.Major != right.Major {
		if left.Major < right.Major {
			return -1
		}
		return 1
	}
	if left.Minor != right.Minor {
		if left.Minor < right.Minor {
			return -1
		}
		return 1
	}
	if left.Patch < right.Patch {
		return -1
	}
	if left.Patch > right.Patch {
		return 1
	}
	return 0
}

// SelectVersion returns the highest candidate satisfying a basic npm-style
// range: exact versions, ^, ~, comparison chains, and *.
func SelectVersion(rangeSpec string, candidates []Version) (Version, error) {
	matching := make([]Version, 0, len(candidates))
	for _, candidate := range candidates {
		if Satisfies(candidate, rangeSpec) {
			matching = append(matching, candidate)
		}
	}
	if len(matching) == 0 {
		return Version{}, fmt.Errorf("no version satisfies %q", rangeSpec)
	}
	sort.Slice(matching, func(i, j int) bool { return CompareVersions(matching[i], matching[j]) > 0 })
	return matching[0], nil
}

func Satisfies(version Version, rangeSpec string) bool {
	rangeSpec = strings.TrimSpace(rangeSpec)
	if rangeSpec == "" || rangeSpec == "*" || rangeSpec == "latest" {
		return true
	}
	for _, alternative := range strings.Split(rangeSpec, "||") {
		if satisfiesChain(version, strings.Fields(strings.TrimSpace(alternative))) {
			return true
		}
	}
	return false
}

func satisfiesChain(version Version, terms []string) bool {
	for _, term := range terms {
		operator := "="
		value := term
		for _, candidate := range []string{">=", "<=", ">", "<", "^", "~", "="} {
			if strings.HasPrefix(term, candidate) {
				operator, value = candidate, strings.TrimSpace(strings.TrimPrefix(term, candidate))
				break
			}
		}
		target, err := ParseVersion(value)
		if err != nil {
			return false
		}
		switch operator {
		case "=":
			if CompareVersions(version, target) != 0 {
				return false
			}
		case ">=":
			if CompareVersions(version, target) < 0 {
				return false
			}
		case "<=":
			if CompareVersions(version, target) > 0 {
				return false
			}
		case ">":
			if CompareVersions(version, target) <= 0 {
				return false
			}
		case "<":
			if CompareVersions(version, target) >= 0 {
				return false
			}
		case "^":
			if CompareVersions(version, target) < 0 || CompareVersions(version, caretUpperBound(target)) >= 0 {
				return false
			}
		case "~":
			if CompareVersions(version, target) < 0 || CompareVersions(version, Version{Major: target.Major, Minor: target.Minor + 1}) >= 0 {
				return false
			}
		}
	}
	return true
}

func caretUpperBound(version Version) Version {
	if version.Major > 0 {
		return Version{Major: version.Major + 1}
	}
	if version.Minor > 0 {
		return Version{Major: 0, Minor: version.Minor + 1}
	}
	return Version{Major: 0, Minor: 0, Patch: version.Patch + 1}
}
