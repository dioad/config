package configcheck_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dioad/config/configcheck"
)

func findingsOf(check func(*configcheck.Collector)) []configcheck.Finding {
	c := configcheck.NewCollector()
	check(c.Field("x"))

	return c.Findings()
}

func TestCheckListenAddress(t *testing.T) {
	t.Parallel()

	for _, ok := range []string{"", ":1080", "0.0.0.0:1080", "localhost:0", "[::1]:8443"} {
		assert.Empty(t, findingsOf(func(c *configcheck.Collector) { configcheck.CheckListenAddress(c, ok) }), ok)
	}
	assert.Equal(t, []configcheck.Finding{{Path: "x", Message: `"1080": address 1080: missing port in address`}},
		findingsOf(func(c *configcheck.Collector) { configcheck.CheckListenAddress(c, "1080") }))
	assert.Equal(t, []configcheck.Finding{{Path: "x", Message: `"localhost:" has no port`}},
		findingsOf(func(c *configcheck.Collector) { configcheck.CheckListenAddress(c, "localhost:") }))
}

func TestCheckURL(t *testing.T) {
	t.Parallel()

	for _, ok := range []string{"", "https://example.com/path", "http://localhost:8080", "tcp://example.com:8119"} {
		assert.Empty(t, findingsOf(func(c *configcheck.Collector) { configcheck.CheckURL(c, ok) }), ok)
	}
	assert.Equal(t, []configcheck.Finding{{Path: "x", Message: `"example.com:8119" must be a URL with a host, such as https://example.com/path`}},
		findingsOf(func(c *configcheck.Collector) { configcheck.CheckURL(c, "example.com:8119") }))
}

func TestCheckHTTPURL(t *testing.T) {
	t.Parallel()

	for _, ok := range []string{"", "https://events.example.com/event", "http://localhost:8120/event"} {
		assert.Empty(t, findingsOf(func(c *configcheck.Collector) { configcheck.CheckHTTPURL(c, ok) }), ok)
	}
	assert.Equal(t, []configcheck.Finding{{Path: "x", Message: `"wss://events.example.com" must be an http or https URL`}},
		findingsOf(func(c *configcheck.Collector) { configcheck.CheckHTTPURL(c, "wss://events.example.com") }))
	assert.Equal(t, []configcheck.Finding{{Path: "x", Message: `"https:///event" has no host`}},
		findingsOf(func(c *configcheck.Collector) { configcheck.CheckHTTPURL(c, "https:///event") }))
}

func TestCheckNonNegative(t *testing.T) {
	t.Parallel()

	assert.Empty(t, findingsOf(func(c *configcheck.Collector) { configcheck.CheckNonNegative(c, 0) }))
	assert.Equal(t, []configcheck.Finding{{Path: "x", Message: "-1s must not be negative"}},
		findingsOf(func(c *configcheck.Collector) { configcheck.CheckNonNegative(c, -time.Second) }))
}

func TestIsLoopbackAddress(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"127.0.0.1:5555":   true,
		"[::1]:5555":       true,
		"localhost:5555":   true,
		"LOCALHOST:5555":   true,
		"0.0.0.0:5555":     false,
		":5555":            false,
		"10.0.0.5:5555":    false,
		"[::]:5555":        false,
		"example.com:5555": false,
		"not-an-address":   false,
		"":                 false,
	}
	for address, want := range tests {
		t.Run(address, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, configcheck.IsLoopbackAddress(address))
		})
	}
}

func TestListenAddressesConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b string
		want bool
	}{
		{"0.0.0.0:8443", "0.0.0.0:8443", true},
		{"127.0.0.1:8443", "127.0.0.1:8443", true},
		{"0.0.0.0:8443", "127.0.0.1:8443", true},
		{":8443", "10.0.0.5:8443", true},
		{"[::]:8443", "127.0.0.1:8443", true},
		{"127.0.0.1:8443", "10.0.0.5:8443", false},
		{"0.0.0.0:8443", "0.0.0.0:8444", false},
		{"127.0.0.1:0", "127.0.0.1:0", false},
		{"not-an-address", "not-an-address", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.a+" "+tt.b, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, configcheck.ListenAddressesConflict(tt.a, tt.b))
			assert.Equal(t, tt.want, configcheck.ListenAddressesConflict(tt.b, tt.a), "the check is symmetric")
		})
	}
}

func ExampleCheckListenAddress() {
	findings := configcheck.NewCollector()
	configcheck.CheckListenAddress(findings.Field("management").Field("listen-address"), "4444")
	configcheck.CheckListenAddress(findings.Field("control").Field("listen-address"), ":8119")

	for _, f := range findings.Findings() {
		fmt.Println(f)
	}
	// Output: management.listen-address: "4444": address 4444: missing port in address
}

func ExampleListenAddressesConflict() {
	fmt.Println(configcheck.ListenAddressesConflict(":8443", "10.0.0.5:8443"))
	fmt.Println(configcheck.ListenAddressesConflict("127.0.0.1:8443", "10.0.0.5:8443"))
	// Output:
	// true
	// false
}
