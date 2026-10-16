# config

Checks for configuration that dioad services, libraries and plugins load:

- `configcheck` decodes configuration strictly, rejecting keys that match no
  field, with the same decode hooks the dioad/cli loader uses, and reports
  every problem as a finding with the path of the field it concerns.

It depends only on the standard library, mapstructure and dioad/util, so a
library or plugin can check configuration without importing a CLI framework.
