

# INTERCEPT OBSERVE 



<br><br>


```sh

Usage:
  intercept observe [flags]

Flags:
      --env-detection        Enable environment detection if no environment is specified
      --environment string   Filter policies that match the specified environment
  -h, --help                 help for observe
      --index string         Index name for ES bulk operations (default "intercept")
      --mode string          Observe mode for path monitoring : first,last,all  (default "last")
      --policy string        Policy file
      --report string        Report Cron Schedule
      --schedule string      Global Cron Schedule
      --tags_all string      Filter policies that match all of the provided tags (comma-separated)
      --tags_any string      Filter policies that match any of the provided tags (comma-separated)

  ```
  ## Feature Flags

### --policy
Load a policy locally or from a remote endpoint
```sh
--policy policies/scan.yml
--policy https://intercept.cc/marketplace/nginx_policy.yml
```
### --environment
Declare the environment to assess the severity level of your policies
```sh
--environment production
# Defaults "all"
```

### --env-detection
Automatically detects the environment variable from common dev paths
```sh
--env-detection 
# Superseeded by --environment
```

### --mode
Observe mode for path monitoring, chose to which monitoring event order to react to. 
```sh
# Options: first || last || all  
# Default: "last" (reacts to the last event triggered for a path)
--mode first
```

### --report
Report Cron Schedule for the cadence of full compliance reports generated
```sh
# should be an interval delta long enough to capture 
# individual policy audit results
--report */50 * * * * *
```

### --schedule
Global Cron Schedule, for policies without individual schedule
```sh
# set a cadence of running the individual policy audits
--schedule */10 * * * * *
```

### --tags-all
Filter policies that match all of the provided tags
Only runs the Audit on policies with ALL the declared tags
```sh
--tags-all security,rbac
```
### --tags-any
Filter policies that match any of the provided tags
```sh
--tags-any security,compliance
```