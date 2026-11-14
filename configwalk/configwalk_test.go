package configwalk_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dioad/config/configwalk"
)

type tlsBlock struct {
	CertFile string `mapstructure:"cert-file"`
}

type embedded struct {
	Shared string `mapstructure:"shared"`
}

type rateLimit struct {
	RPS float64 `mapstructure:"rate-limit-rps"`
}

type node struct {
	Name string `mapstructure:"name"`
	Next *node  `mapstructure:"next"`
}

type sample struct {
	embedded

	ListenAddress string              `mapstructure:"listen-address"`
	Plain         string              // no tag: keyed by its lower-cased Go name
	Drain         time.Duration       `mapstructure:"drain-delay"`
	Origins       []string            `mapstructure:"cors-allowed-origins"`
	TLS           tlsBlock            `mapstructure:"tls-server"`
	Limit         rateLimit           `mapstructure:"limit,squash"`
	Skipped       string              `mapstructure:"-"`
	hidden        string              //nolint:unused // proves an unexported field is not a key
	Labels        map[string]tlsBlock `mapstructure:"labels"`
}

func keys(fields []configwalk.Field) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, f.Key)
	}

	return out
}

func TestFields_ReadsKeysTheWayTheLoaderDoes(t *testing.T) {
	t.Parallel()

	got := configwalk.Fields(reflect.TypeFor[sample]())

	assert.Equal(t, []string{"shared", "listen-address", "plain", "drain-delay", "cors-allowed-origins", "tls-server", "rate-limit-rps", "labels"}, keys(got),
		"an embedded struct and a squashed one are inlined; ignored and unexported fields are not keys")
	assert.Equal(t, reflect.TypeFor[rateLimit](), got[6].Owner, "an inlined field knows its own struct")
}

func TestWalk_VisitsEachReachableStructOnce(t *testing.T) {
	t.Parallel()
	var visited []string

	configwalk.Walk(reflect.TypeFor[sample](), func(t reflect.Type, _ []configwalk.Field) { visited = append(visited, t.Name()) })

	assert.ElementsMatch(t, []string{"sample", "embedded", "tlsBlock", "rateLimit"}, visited, "through fields, squashed structs and map values")

	visited = nil
	configwalk.Walk(reflect.TypeFor[node](), func(t reflect.Type, _ []configwalk.Field) { visited = append(visited, t.Name()) })
	assert.Equal(t, []string{"node"}, visited, "a struct that refers to itself is visited once")
}

func TestSquashedFields_ListsNamedSquashedStructs(t *testing.T) {
	t.Parallel()

	got := configwalk.SquashedFields(reflect.TypeFor[sample]())

	require.Len(t, got, 1)
	assert.Equal(t, "Limit", got[0].Name)
}

func TestDefaults_RecordsWhatTheConstructorSets(t *testing.T) {
	t.Parallel()
	prototype := &sample{ListenAddress: ":8080", Drain: 5 * time.Second, Origins: []string{"a", "b"}, TLS: tlsBlock{CertFile: "ignored"}, Limit: rateLimit{RPS: 2}}

	assert.Equal(t, map[string]any{
		"listen-address":       ":8080",
		"drain-delay":          "5s",
		"cors-allowed-origins": []any{"a", "b"},
		"rate-limit-rps":       2.0,
	}, configwalk.Defaults(prototype), "a squashed struct's defaults are the parent's; a nested struct records none")
}

func TestDefaults_OfANonStructIsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, configwalk.Defaults(42))
	assert.Nil(t, configwalk.Defaults((*sample)(nil)))
}

func TestKeyNamingProblems_FindsBadKeysAtAnyDepth(t *testing.T) {
	t.Parallel()

	// The keys are wrong on purpose.
	type leaf struct {
		Good   string `json:"good"                      mapstructure:"good"`
		Bad    string `json:"bad_config"`
		Suffix string `mapstructure:"tls-server-config"`
		Camel  string `mapstructure:"listenAddress"`  //nolint:tagliatelle // deliberately bad
		Snake  string `mapstructure:"listen_address"` //nolint:tagliatelle // deliberately bad
	}
	type root struct {
		Items []map[string]*leaf `mapstructure:"items"`
	}

	problems := configwalk.KeyNamingProblems(reflect.TypeFor[root]())

	got := make([]string, 0, len(problems))
	for _, p := range problems {
		got = append(got, p.String())
	}
	assert.ElementsMatch(t, []string{
		`configwalk_test.leaf.Bad: bad_config: ends in "config"; drop the suffix`,
		`configwalk_test.leaf.Suffix: tls-server-config: ends in "config"; drop the suffix`,
		`configwalk_test.leaf.Camel: listenAddress: is not kebab-case`,
		`configwalk_test.leaf.Snake: listen_address: is not kebab-case`,
	}, got)
}

func TestKeyNamingProblems_AcceptsAWellNamedConfig(t *testing.T) {
	t.Parallel()

	assert.Empty(t, configwalk.KeyNamingProblems(reflect.TypeFor[sample]()))
}

func ExampleFields() {
	type config struct {
		ListenAddress string `mapstructure:"listen-address"`
		Limit         struct {
			RPS float64 `mapstructure:"rate-limit-rps"`
		} `mapstructure:"limit,squash"`
	}

	for _, f := range configwalk.Fields(reflect.TypeFor[config]()) {
		fmt.Println(f.Key)
	}
	// Output:
	// listen-address
	// rate-limit-rps
}

func ExampleKeyNamingProblems() {
	type config struct {
		TLS    string `mapstructure:"tls-server-config"`
		Listen string `mapstructure:"listen-address"`
	}

	for _, p := range configwalk.KeyNamingProblems(reflect.TypeFor[config]()) {
		fmt.Println(p.Key, "-", p.Reason)
	}
	// Output: tls-server-config - ends in "config"; drop the suffix
}
