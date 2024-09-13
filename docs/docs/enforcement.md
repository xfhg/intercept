

# Policy Enforcement Levels



Enforcement levels are a first-class concept in Intercept, allowing compliant/non-compliant (pass/fail) behavior to be associated separately from the policy logic. This enables any policy to be configured as a warning, allow exceptions, or be absolutely mandatory. These levels can be coupled to environments, allowing different uses of the same policy to have distinct enforcement levels per environment.

## Enforcement Levels

<br>

| Fatal | Exceptions | Confidence | SARIF Level | Description |
|-------|------------|------------|-------------|-------------|
| true  | false      | high       | error       | Highest-confidence, fatal issue with no exceptions |
| true  | true       | high       | error       | High-confidence, potentially fatal issue with exceptions |
| false | false      | high       | error       | High-confidence, non-fatal issue with no exceptions |
| false | true       | high       | warning     | High-confidence, non-fatal issue with exceptions |
| false | false      | low        | warning     | Low-confidence, non-fatal issue with no exceptions |
| false | true       | low        | note        | Lowest-confidence, non-fatal issue with exceptions |
| false | true       | info       | none        | Informational finding |



## SARIF Level


- **error:** A serious issue that very likely indicates a problem in the code.
- **warning:** A potential issue that may or may not indicate a problem in the code.
- **note:** An informational finding that doesn't necessarily indicate a problem.
- **none:** A finding that doesn't have a severity associated with it.

::: warning WIP
This document is a work in progress. Please check back for updates.
:::