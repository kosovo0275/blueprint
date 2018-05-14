package proptools

import "strings"

// NinjaEscape takes a slice of strings that may
// contain characters that are meaningful to ninja ($),
// and escapes each string so they will be passed to
// bash. It is not necessary on input, output, or
// dependency names, those are handled by
// ModuleContext.Build. It is generally required on
// strings from properties in Blueprint files that are
// used as Args to ModuleContext.Build. A new slice
// containing the escaped strings is returned.
func NinjaEscape(slice []string) []string {
	slice = append([]string(nil), slice...)
	for i, s := range slice {
		slice[i] = ninjaEscaper.Replace(s)
	}
	return slice
}

var ninjaEscaper = strings.NewReplacer(
	"$", "$$")

// ShellEscape takes a slice of strings that may
// contain characters that are meaningful to bash and
// escapes if necessary by wrapping them in single
// quotes, and replacing internal single quotes with
// '\'' (one single quote to end the quoting, a
// shell-escaped single quote to insert a real single
// quote, and then a single quote to restart quoting.
// A new slice containing the escaped strings is
// returned.
func ShellEscape(slice []string) []string {
	shellUnsafeChar := func(r rune) bool {
		switch {
		case 'A' <= r && r <= 'Z',
			'a' <= r && r <= 'z',
			'0' <= r && r <= '9',
			r == '_',
			r == '+',
			r == '-',
			r == '=',
			r == '.',
			r == ',',
			r == '/',
			r == ' ':
			return false
		default:
			return true
		}
	}

	slice = append([]string(nil), slice...)

	for i, s := range slice {
		if strings.IndexFunc(s, shellUnsafeChar) == -1 {
			// No escaping necessary
			continue
		}

		slice[i] = `'` + singleQuoteReplacer.Replace(s) + `'`
	}
	return slice

}

func NinjaAndShellEscape(slice []string) []string {
	return ShellEscape(NinjaEscape(slice))
}

var singleQuoteReplacer = strings.NewReplacer(`'`, `'\''`)
