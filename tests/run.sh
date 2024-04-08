#!/bin/sh

intercept config -r 
intercept config -a /app/examples/policy/filetype.yaml

intercept yml -t /app/examples/target/
intercept toml -t /app/examples/target/
intercept json -t /app/examples/target/

intercept config -r 
intercept config -a /app/examples/policy/assure.yaml
intercept config -a /app/examples/policy/assure.yaml

intercept audit -t /app/examples/target -i "AWS" -b "false"
intercept scan -t /app/examples/target -i "AWS" -b "false"
intercept collect -t /app/examples/target -i "AWS" -b "false"

ls -la

cat intercept.audit.sarif.json