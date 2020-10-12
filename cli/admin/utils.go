package admin

import (
	"bytes"
	"fmt"
	"strings"
)

type NameUsage struct {
	Name  string
	Usage string
}

type NameUsages []NameUsage

func (nameUsages NameUsages) Len() int {
	return len(nameUsages)
}

func swap(s []string, i int, j int) {
	s[i], s[j] = s[j], s[i]
}

func (nameUsages NameUsages) Swap(i int, j int) {
	nameUsages[i], nameUsages[j] = nameUsages[j], nameUsages[i]
}

func (nameUsages NameUsages) Less(i int, j int) bool {
	return nameUsages[i].Name < nameUsages[j].Name
}

func (nameUsages NameUsages) Format() string {
	resLines := []string{}
	const sepStr = "\x00"

	maxNameLen := 0

	for _, nameUsage := range nameUsages {
		if len(nameUsage.Name) > maxNameLen {
			maxNameLen = len(nameUsage.Name)
		}

		resLines = append(resLines, nameUsage.Name+sepStr+nameUsage.Usage)
	}

	buf := new(bytes.Buffer)

	for _, line := range resLines {
		sidx := strings.Index(line, sepStr)
		spacing := strings.Repeat(" ", maxNameLen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(buf, line[:sidx], spacing, wrap(maxNameLen+2, 0, line[sidx+1:]))
	}

	return buf.String()
}

func convertToMap(raw interface{}) (map[interface{}]interface{}, error) {
	rawMap, ok := raw.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Isn't a map: %#v", raw)
	}

	return rawMap, nil
}

func getStrValueFromRawMap(mapRaw map[interface{}]interface{}, key string) (string, error) {
	valueRaw, found := mapRaw[key]
	if !found {
		return "", fmt.Errorf("Key %q isn't found", key)
	}

	valueStr, ok := valueRaw.(string)
	if !ok {
		return "", fmt.Errorf("Value %q isn't a string: %#v", key, valueRaw)
	}

	return valueStr, nil
}

// See https://github.com/spf13/pflag/blob/81378bbcd8a1005f72b1e8d7579e5dd7b2d612ab/flag.go#L612

// Splits the string `s` on whitespace into an initial substring up to
// `i` runes in length and the remainder. Will go `slop` over `i` if
// that encompasses the entire string (which allows the caller to
// avoid short orphan words on the final line).
func wrapN(i, slop int, s string) (string, string) {
	if i+slop > len(s) {
		return s, ""
	}

	w := strings.LastIndexAny(s[:i], " \t\n")
	if w <= 0 {
		return s, ""
	}
	nlPos := strings.LastIndex(s[:i], "\n")
	if nlPos > 0 && nlPos < w {
		return s[:nlPos], s[nlPos+1:]
	}
	return s[:w], s[w+1:]
}

// See https://github.com/spf13/pflag/blob/81378bbcd8a1005f72b1e8d7579e5dd7b2d612ab/flag.go#L632

// Wraps the string `s` to a maximum width `w` with leading indent
// `i`. The first line is not indented (this is assumed to be done by
// caller). Pass `w` == 0 to do no wrapping
func wrap(i, w int, s string) string {
	if w == 0 {
		return strings.Replace(s, "\n", "\n"+strings.Repeat(" ", i), -1)
	}

	// space between indent i and end of line width w into which
	// we should wrap the text.
	wrap := w - i

	var r, l string

	// Not enough space for sensible wrapping. Wrap as a block on
	// the next line instead.
	if wrap < 24 {
		i = 16
		wrap = w - i
		r += "\n" + strings.Repeat(" ", i)
	}
	// If still not enough space then don't even try to wrap.
	if wrap < 24 {
		return strings.Replace(s, "\n", r, -1)
	}

	// Try to avoid short orphan words on the final line, by
	// allowing wrapN to go a bit over if that would fit in the
	// remainder of the line.
	slop := 5
	wrap = wrap - slop

	// Handle first line, which is indented by the caller (or the
	// special case above)
	l, s = wrapN(wrap, slop, s)
	r = r + strings.Replace(l, "\n", "\n"+strings.Repeat(" ", i), -1)

	// Now wrap the rest
	for s != "" {
		var t string

		t, s = wrapN(wrap, slop, s)
		r = r + "\n" + strings.Repeat(" ", i) + strings.Replace(t, "\n", "\n"+strings.Repeat(" ", i), -1)
	}

	return r

}
