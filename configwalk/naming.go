package configwalk

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var kebabCase = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// KeyProblem is a config key that breaks the naming rules.
type KeyProblem struct {
	// Field names the field as "Type.Field".
	Field string
	// Key is the offending key.
	Key string
	// Reason says which rule it breaks.
	Reason string
}

// String returns the problem as "Type.Field: key: reason".
func (p KeyProblem) String() string {
	return fmt.Sprintf("%s: %s: %s", p.Field, p.Key, p.Reason)
}

// KeyNamingProblems checks the keys of every struct reachable from root against
// the naming rules Dioad config files follow, and returns each key that breaks
// one:
//
//   - a key must not end in "config", since being in a config already says so:
//     "tls-server", not "tls-server-config". This applies to the mapstructure
//     key and to the json name of the same field, which names it in an API.
//   - a mapstructure key is kebab-case: lower-case words joined by hyphens.
//
// A key may be exempt for a reason; filter the result in the test that calls
// this, keeping the reason beside the exemption.
func KeyNamingProblems(root reflect.Type) []KeyProblem {
	var problems []KeyProblem
	Walk(root, func(t reflect.Type, _ []Field) {
		for i := range t.NumField() {
			f := t.Field(i)
			if IsIgnored(f) {
				continue
			}
			where := t.String() + "." + f.Name

			if key, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ","); key != "" {
				if !kebabCase.MatchString(key) {
					problems = append(problems, KeyProblem{where, key, "is not kebab-case"})
				}
				if endsInConfig(key) {
					problems = append(problems, KeyProblem{where, key, `ends in "config"; drop the suffix`})
				}
			}
			if key, _, _ := strings.Cut(f.Tag.Get("json"), ","); endsInConfig(key) {
				problems = append(problems, KeyProblem{where, key, `ends in "config"; drop the suffix`})
			}
		}
	})

	return problems
}

func endsInConfig(key string) bool {
	return strings.HasSuffix(key, "-config") || strings.HasSuffix(key, "_config")
}
