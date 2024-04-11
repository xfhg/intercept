
# Configuration & Flags


::: tip
All flags under the same instruction can be combined
:::


<br>

- SHA256 Hash for configuration file 
```sh
intercept config -a examples/yourpolicies.yaml -k 201c8fe265808374f3354768410401216632b9f2f68f9b15c85403da75327396
```
- Download of configuration file
```sh
intercept config -a https://xxx.com/artifact/policy.yaml -k 201c8fe26580(...)
```
- Enviroment enforcement (check policy enforcement levels)
```sh
intercept audit -t examples/target -e "development"
```
- Rule Tag filter
```sh
intercept audit -t examples/target -i "AWS,OWASP"
```
- Run only the API type checks (but with tags)
```sh
intercept api -i "AWS,OWASP"
```


- Run only the YML type checks
```sh
intercept yml -t examples/target 
```
- Disable pipeline break
```sh
intercept audit -t examples/target -b false
```
- Accept No exceptions 
```sh
intercept audit -t examples/target -x
```
- Ignoring files and folders
```
use a .ignore file or your own .gitignore
```