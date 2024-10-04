

# INTERCEPT AUDIT 

<br><br>


```sh
Usage:
  intercept audit [flags]

Flags:
      --checksum string      Policy SHA256 expected checksum
      --env-detection        Enable environment detection if no environment is specified
  -e, --environment string   Filter policies that match the specified environment
  -h, --help                 help for audit
  -p, --policy string        Policy <FILEPATH> or <URL>
      --tags-all string      Filter policies that match all of the provided tags (comma-separated)
  -f, --tags-any string      Filter policies that match any of the provided tags (comma-separated)
  -t, --target string        Target directory to audit


  ```
## Feature Flags

### --policy
Load a policy locally or from a remote endpoint
```sh
--policy policies/scan.yml
--policy https://intercept.cc/marketplace/nginx_policy.yml
```
### --checksum
Expected SHA256 Checksum of the policy file
```sh
--checksum a3717edde60a3f80fd6c401a666ca1f9b0ea6542b7834009452e2439d8951307
```
### --target
Base target directory to audit
```sh
# Policies like SCAN , ASSURE , REGO , etc 
# need a target path to look/filter for target files
--target targets/
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
### --tags-all
Only runs the Audit on policies with ALL the declared tags
```sh
--tags-all security,rbac
```