// Package util contains some util functions used by jiebago.
package util

import "regexp"

/*
RegexpSplit split slices s into substrings separated by the expression and
returns a slice of the substrings between those expression matches.
If capturing parentheses are used in expression, then the text of all groups
in the expression are also returned as part of the resulting slice.

This function acts consistent with Python's re.split function.
*/
func RegexpSplit(re *regexp.Regexp, s string, n int) []string {
	if n == 0 {
		return nil
	}

	if len(re.String()) > 0 && len(s) == 0 {
		return []string{""}
	}

	var matches [][]int
	if len(re.SubexpNames()) > 1 {
		matches = re.FindAllStringSubmatchIndex(s, n)
	} else {
		matches = re.FindAllStringIndex(s, n)
	}
	strings := make([]string, 0, len(matches))

	beg := 0
	end := 0
	for _, match := range matches {
		if n > 0 && len(strings) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			strings = append(strings, s[beg:end])
		}
		beg = match[1]
		if len(re.SubexpNames()) > 1 {
			strings = append(strings, s[match[0]:match[1]])
		}
	}

	if end != len(s) {
		strings = append(strings, s[beg:])
	}

	return strings
}
