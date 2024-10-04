

# INTERCEPT Global Feature Flags

<br><br>

::: tip HINT
Most flags can be declared on the policy file
:::


```sh
Usage:
  intercept [command]

Available Commands:
  audit       Run an optimized audit through all loaded policies
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  observe     Observe and trigger realtime policies based on schedules or active path monitoring
  sys         Test intercept embedded core binaries
  version     Print the build info of intercept

Flags:
      --debug                Enable extra dev debug output
      --experimental         Enables unreleased experimental features
  -h, --help                 help for intercept
      --log-type string      Compliance Log types (can be a list) : MINIMAL,RESULTS,POLICY,REPORT (default "RESULTS")
      --nolog                Disables all loggging
  -o, --output-dir string    directory to write output files
      --output-type string   Output types (can be a list) : SARIF,LOG (default "SARIF")
      --silent               Enables log to file intercept.log
  -v, --verbose count        increase verbosity level
```

## Configuration Flags

### -v 
Verbosity Level
```sh
# Default : Disabled
# Levels of verbose output : DEBUG / INFO / WARN / ERROR / FATAL
-v
-vv
-vvv
-vvvv
```
### --debug
Enable extra dev debug output
```sh
# Default : false
# Should be used with -vvvv
# Enable extra dev debug output :

2024-10-... WRN DEBUG OUTPUT ENABLED
2024-10-... WRN DEBUG OUTPUT ENABLED - Output can print sensitive data
2024-10-... WRN DEBUG OUTPUT ENABLED

```

### -o 
Declare a directory to write output files (logs +debug + reports)
```sh
# Default ./

-o _reports/

# tree _reports/
_reports/
├── _debug      # dev help output
├── _patched    # automated fixed files
├── _sarif      # intermediary individual results (cache)
├── _status     # compliance reports 
│   ├── 20241004T171050Z_intercept_2myvsh.sarif.json
│   └── 20241004T171100Z_intercept_2myvsh.sarif.json
├── log_minimal_2myvsh.log
├── log_policy_2myvsh.log
├── log_report_2myvsh.log
└── log_results_2myvsh.log
```

### --output-type
Output types (can be a list) : SARIF,LOG (default "SARIF")
```sh
--output-type LOG
# to be used with --log-type
```

### --log-type
Compliance Log types (can be a list) 
```sh
--log-type minimal,results,policy,report
# Experiment with all
# MINIMAL,RESULTS,POLICY,REPORT (default "RESULTS")
```

### --nolog
Disables all intercept logging and output (not the compliance reporting)

### --silent
Redirects operational intercept log to file intercept.log

