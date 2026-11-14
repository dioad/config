package configwalk

// The patterns of the strings the config loader's weakly typed decoding
// converts to a number or a duration, for a generator that writes the schema of
// a scalar field. mapstructure parses an integer string with strconv.ParseInt
// in base 0, so it also accepts forms such as 0x1f and 1_000; these patterns
// accept decimal integers only. An empty string decodes to zero.
const (
	IntegerPattern  = `^([-+]?[0-9]+)?$`
	UnsignedPattern = `^(\+?[0-9]+)?$`
	NumberPattern   = `^([-+]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][-+]?[0-9]+)?)?$`
	// DurationPattern matches what time.ParseDuration accepts.
	DurationPattern = `^[-+]?(0|(([0-9]+(\.[0-9]*)?|\.[0-9]+)(ns|us|µs|μs|ms|s|m|h))+)$`
)

// BoolStrings are the strings strconv.ParseBool accepts, which weakly typed
// decoding converts to a bool, and "", which it decodes to false.
var BoolStrings = []any{"", "1", "t", "T", "TRUE", "true", "True", "0", "f", "F", "FALSE", "false", "False"}
