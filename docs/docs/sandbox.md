# Sandbox Playground

::: tip START HERE
Free to play, no hassle intercept sandbox :

Check an insecure nginx.conf in less than 20 miliseconds
:::

<br><br>

## 1. Our friends at Gitpod will host you 

Gitpod offers a convenient way to explore Intercept in a cloud environment

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)

## 2. Build & Experiment :

Once in the Gitpod environment, you can build Intercept and start experimenting:

```
make build && cp release/intercept playground/intercept
```

## 3. Explore the playground folder

```sh{3,11}
playground/
│
├─ policies/
│  ├─ ..yaml
│  ├─ ..yaml        example INTERCEPT policies for all policy types
│  └─ ..yaml
├─ runtime/
│  ├─ ..yaml
│  ├─ ..yaml        example RUNTIME partial policies 
│  └─ ..yaml
└─ targets/
   ├─ ..json
   ├─ ..toml        example target configuration files
   └─ ..ini
```
## 🛰️ INTERCEPT AUDIT

## 4. Check for an insecure nginx.conf in less than 20 miliseconds

### our playground example

```sh{5}
cd playground

./intercept audit \
  --policy policies/nginx_insecure.yaml \
  --target targets/ \
  -vvvv \
  -o _playground_nginx
```

### the gitpod nginx.conf

```sh{3}
./intercept audit \
  --policy policies/nginx_insecure.yaml \
  --target /etc/nginx/ \
  -vvvv \
  -o _gitpod_nginx
```


## 5. Check the compliance report

```
find it inside the output folders (_gitpod_nginx and _playground_nginx)
```

```json
 "invocations": [
        {
          "executionSuccessful": true,
          "commandLine": "./intercept audit --policy policies/nginx_insecure.yaml --target /etc/nginx/ -vvvv -o _gitpod_nginx",
          "properties": {
            "end-time": "2024-09-18T06:13:53Z",
            "execution-time-ms": "15", // [!code focus]
            "report-compliant": "false",
            "report-status": "non-compliant", // [!code focus]
            "report-timestamp": "2024-09-18T06:13:53Z",
            "start-time": "2024-09-18T06:13:53Z"
          }
        }
]
```

::: info
Great, let's automate it and send the report to the CIO
:::

## 🛰️ INTERCEPT OBSERVE

## 6. Setup a monitor for changes in nginx


## 7. Schedule the creation of a report 


## 8. Setup your own integration webhooks to receive scan results and compliance reports


## 9. EXTRA Validate CUE Lang Schemas and REGO Policies

To validate your CUE Lang Schemas or REGO policies, you can use these online tools:


- [CUE Sandbox](https://cuelang.org/play/#cue@export@cue)
    - Use this sandbox to test and refine your CUE Lang schemas.

- [REGO Sandbox](https://play.openpolicyagent.org/p/ZWGVA8oCSE)
    - This sandbox allows you to experiment with and validate your REGO policies.


These tools provide an excellent way to ensure your schemas and policies are correct before adding them in your Intercept Policies.

## 10. Run all these examples with intercept container

```sh 
cd playground
./intercept audit --policy policies/test_scan.yaml --target targets -vvv -o _my_first_run
```