# XXX Type Policies

SCAN-type policies are designed to identify and flag patterns or a collection of known patterns within your target codebase that should ideally be absent. These undesirable patterns may include, but are not limited to, exposed API keys, secrets, overly permissive CIDR ranges in security groups, improperly enabled configuration parameters, or even instances of code, script, or proxy definition misuse that could pose risks to your environment.

This proactive approach allows developers and security professionals to systematically weed out configurations or code snippets that contradict best practices or security guidelines, thereby enhancing the overall security posture and compliance of the codebase.

With an extensive library of over 1,500 predefined patterns available for selection, our tool offers a comprehensive means to safeguard your codebase against common pitfalls and security vulnerabilities. To illustrate, consider the following simplified example:



::: tip
To easily run the examples, clone our repository to have access to a playground
:::



# Playground


::: info
All rule types can be filtered by a combination of TAGS, ENVIRONMENT name and their own ENFORCEMENT levels. Make sure to explore it.
:::

## XXX policy schema

```yaml{7,12-18}
.
├─ index.md
├─ foo
│  ├─ index.md
│  ├─ one.md // [!code focus]
│  └─ two.md
└─ bar
   ├─ index.md
   ├─ three.md // [!code focus]
   └─ four.md

```

## Minimal Policy
```yaml{7,12-18}
- name: Private key committed in code
    id: 100
    description: Private key committed to code version control
    error: This violation immediately blocks your code deployment
    solution: Revoke the detected key
    tags: KEY,CERT
    type: scan
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

```

## Running the AUDIT

## Checking the results

## Tuning the AUDIT


## Full Policy
```yaml{7,12-18}
- name: Private key committed in code
    id: 100
    description: Private key committed to code version control
    error: This violation immediately blocks your code deployment
    solution: Revoke the detected key
    tags: KEY,CERT
    type: scan
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

```



```sh
intercept config -r 
intercept config -a /app/examples/policy/assure.yaml
intercept scan -t /app/examples/target -i "AWS"
cat intercept.audit.sarif.json
```






```yaml{7,12-18}
- name: Private key committed in code
    id: 100
    description: Private key committed to code version control
    error: This violation immediately blocks your code deployment
    solution: Revoke the detected key
    tags: KEY,CERT
    type: scan
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

```



### SARIF Output

```json{8,21,22}
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    ... REDACTED
      "results": [
        {
          "ruleId": "intercept.cc.scan.policy.100: PRIVATE KEY COMMITTED IN CODE",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "Private key committed to code version control"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/long.code"
                },
                "region": {
                  "startLine": 10927,
                  "endLine": 10927,
                  "snippet": {
                    "text": "-----BEGIN PGP PRIVATE KEY BLOCK-----"
                  }
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## Console Output

```sh
├ SCAN Rule # 100
│ Rule name :  Private key committed in code
│ Rule description :  Private key committed to code version control
│ Impacted Env :  all
│ Confidence :  high
│ Tags :  KEY,CERT,AWS
│ 
    /app/examples/target/long.code
    10928:-----BEGIN PGP PRIVATE KEY BLOCK-----
│
│ FATAL : 
│  This violation immediately blocks your code deployment
│
│
│ Rule :  Private key committed in code
│ Target Environment :  all
│ Suggested Solution :  Revoke the detected key
│
│ 
│ 

```


## Run it

```sh
docker pull ghcr.io/xfhg/intercept:latest

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -a examples/policy/assure.yaml

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept scan -t examples/target
```