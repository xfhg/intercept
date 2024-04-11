# Manual Quick Start

1. Grab the latest [RELEASE](https://github.com/xfhg/intercept/releases) of intercept bundle for your platform 

```
core-intercept-rg-x86_64-darwin.zip
core-intercept-rg-x86_64-linux.zip
core-intercept-rg-x86_64-windows.zip
```

2. Make sure you have the latest setup

```sh
intercept system --update
```

3. Load some [EXAMPLE](https://github.com/xfhg/intercept/tree/master/examples) policies and target files

```
start with the minimal.yaml
```

4. Configure intercept

```sh
intercept config -r
intercept config -a examples/policy/minimal.yaml
```

5. Audit the target folder

```sh
intercept audit -t examples/target
```

6. Check the different output flavours

```yml
- stdout human readable console report
- individual json rule output with detailed matches
- all findings compiled into intercept.output.json
- fully compliant SARIF output into intercept.sarif.json
- all SHA256 of the scanned files into intercept.scannedSHA256.json
```

7. Tune the scan with extra flags like ENVIRONMENT or TAGS filter

```sh
intercept audit -t examples/target -e "development" -i "AWS"
```
