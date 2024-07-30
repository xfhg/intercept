#!/bin/sh

# intercept config -r 
# intercept config -a /app/examples/policy/filetype.yaml
# #

# intercept yml -t /app/examples/target/
# cat intercept.yml.sarif.json

# intercept toml -t /app/examples/target/
# cat intercept.toml.sarif.json

# intercept json -t /app/examples/target/
# cat intercept.json.sarif.json

##################################

intercept config -r 
intercept config -a /app/examples/policy/filetype.yaml
# #

intercept ini -t /app/examples/target
# cat intercept.audit.sarif.json

# intercept scan -t /app/examples/target -i "AWS" -b "false"
# cat intercept.audit.sarif.json

# intercept assure -t /app/examples/target -i "AWS" -b "false"
# ls -la
# cat intercept.audit.sarif.json

##################################

# intercept config -r 
# intercept config -a /app/examples/policy/rego.yaml
# #

# intercept rego -t /app/examples/target/rego 
# cat intercept.rego.sarif.json

##################################

# intercept config -r 
# intercept config -a /app/examples/policy/api.yaml
# #
# export INTERCEPT_BAUTH=user:pass

# intercept api 
# cat intercept.api.full.sarif.json
