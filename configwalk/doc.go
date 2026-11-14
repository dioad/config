// Package configwalk reads config structs the way the config loader does, so
// the things generated from them (a JSON Schema, an example file, a reference
// page) and the tests over them agree with what is decoded.
//
// A struct's keys are its fields' mapstructure names, lower-cased as viper
// lower-cases keys; a field with no name uses its Go name. mapstructure reads
// the fields of a struct tagged ",squash" from the parent's keys, named field
// or embedded, and an embedded struct with no mapstructure tag likewise, so
// Fields inlines both. The package depends only on the standard library.
package configwalk
