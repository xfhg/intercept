# Sandbox Playground


<br><br>

### 1. Our friends at Gitpod will host you 

Gitpod offers a convenient way to explore Intercept in a cloud environment

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)

### 2. Build & Experiment :

Once in the Gitpod environment, you can build Intercept and start experimenting:

```
make build && cp release/intercept playground/intercept
```

### 3. Explore the playground folder

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

### 4. Validate CUE Lang Schemas and REGO Policies

To validate your CUE Lang Schemas or REGO policies, you can use these online tools:


- [CUE Sandbox](https://cuelang.org/play/#cue@export@cue)
    - Use this sandbox to test and refine your CUE Lang schemas.




- [REGO Sandbox](https://play.openpolicyagent.org/p/ZWGVA8oCSE)
    - This sandbox allows you to experiment with and validate your REGO policies.


These tools provide an excellent way to ensure your schemas and policies are correct before adding them in your Intercept Policies.

### 5. Run Intercept with some playground examples

Try some example policies before building your own

```sh 
cd playground
./intercept audit --policy policies/test_scan.yaml --target targets -vvv -o _my_first_run
```