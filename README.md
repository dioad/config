# config

Checks for configuration that dioad services, libraries and plugins load:

- `configcheck` decodes configuration strictly, rejecting keys that match no
  field, with the same decode hooks the dioad/cli loader uses, and reports
  every problem as a finding with the path of the field it concerns. It also
  holds the checks services repeat for listen addresses, URLs and durations
  (`CheckListenAddress`, `IsLoopbackAddress`, `ListenAddressesConflict`,
  `CheckURL`, `CheckHTTPURL`, `CheckNonNegative`).
- `configwalk` reads config structs the way the loader does: keys from
  `mapstructure` tags with squashed and embedded structs inlined, a walk over
  every reachable struct, the defaults a constructor sets, the patterns of
  weakly typed scalars, and `KeyNamingProblems`, a test helper that enforces
  the key naming rules (kebab-case, no `config` suffix). A schema or example
  generator, and the tests over a config type, build on it.

It depends only on the standard library, mapstructure and dioad/util, so a
library or plugin can check configuration without importing a CLI framework.
