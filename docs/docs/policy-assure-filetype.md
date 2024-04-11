# ASSURE-FILETYPE Type Policies

ASSURE-type policies serve as proactive compliance tools within your codebase or configuration settings, distinctly contrasting with the reactive nature of SCAN-type policies. 

Instead of searching for known issues or vulnerabilities, ASSURE policies help in affirming the presence of specific, desirable patterns within your target codebase. These patterns are pivotal in ensuring that your codebase adheres to compliance standards or configuration requirements.

This approach shifts the focus from merely identifying and rectifying problems to actively validating and ensuring the desired state of system configurations, thereby fostering a more secure and compliant infrastructure.


::: tip
Filter the target data by filetype and apply/assure CUE Lang schemas directly from the policies
:::

## CUE Lang

Defining the patterns for ASSURE-type policies using CUE Language (CUE Lang) schemas introduces a new level of precision, flexibility, and efficiency in policy specification. CUE Lang, a constraint-based configuration language, offers several distinct advantages for defining these patterns:

- Enhanced Precision and Clarity

CUE Lang's schema-based approach allows for the detailed specification of patterns with clear, unambiguous definitions. This precision ensures that policies are applied consistently, reducing the risk of misinterpretation or errors in policy enforcement. By explicitly outlining the structure and constraints of desired configurations, CUE Lang schemas make it easier to define complex compliance requirements in a way that is both human-readable and machine-enforceable.

- Reusability and Modularity

CUE Lang promotes reusability and modularity through its schema definitions. Patterns defined for one project can be easily adapted or extended for use in others, without the need for duplication of effort. This modularity not only speeds up the policy creation process but also ensures consistency across different projects and teams within an organization.

- Scalability and Evolution

As organizational needs evolve, so too can the ASSURE policies defined using CUE Lang. The language's flexible schema definitions enable policies to be updated or expanded with minimal effort, ensuring that compliance standards can keep pace with changing requirements, technologies, and industry best practices.

- Enhanced Error Detection and Correction

CUE Lang's constraint-based approach not only helps in defining what is correct but also aids in identifying and correcting deviations from the defined policies. When applied to ASSURE-type policies, this means that any discrepancies or omissions in the target codebase or configurations can be precisely pinpointed and addressed, enhancing the overall security and compliance posture.



## Example

### The setup

```sh{2,4}
intercept config -r 
intercept config -a /app/examples/policy/filetype.yaml

intercept yml -t /app/examples/target/
cat intercept.yml.sarif.json

intercept toml -t /app/examples/target/
cat intercept.toml.sarif.json

intercept json -t /app/examples/target/
cat intercept.json.sarif.json

```
::: info
All rule types can be filtered by a combination of TAGS, ENVIRONMENT name and their own ENFORCEMENT levels. Make sure to explore it.
:::


### The Policy

```yaml{7}

 - name: YML ASSURE - ingress enabled
    id: 151
    description: Assure ingress enabled
    error: Misconfiguration or omission is fatal
    tags: INGRESS
    type: yml
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    yml_filepattern: "^development-\\d+\\.yml$"
    yml_structure: |
      ingress: {enabled: true}
      ...

  - name: YML ASSURE - ingress disabled
    id: 152
    description: Assure ingress disabled
    error: Misconfiguration or omission is fatal
    tags: INGRESS
    type: yml
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    yml_filepattern: "^development-\\d+\\.yml$"
    yml_structure: |
      ingress: {enabled: false}
      ...

  - name: YML ASSURE - memory limit
    id: 153
    description: Assure resources
    error: Misconfiguration or omission is fatal
    tags: MEMORY
    type: yml
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    yml_filepattern: "^development-\\d+\\.yml$"
    yml_structure: |
      logging: { level: "info" }
      resources: {limits: { memory: "128Mi" }}
      ...

  - name: TOML ASSURE - database port
    id: 133
    description: Assure database port
    error: Misconfiguration or omission is fatal
    tags: DB
    type: toml
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    toml_filepattern: "^development-\\d+\\.toml$"
    toml_structure: |
      database: {port: 5432 }
      logging: {level : "info"}
      ...

  - name: TOML ASSURE - missing database port
    id: 134
    description: Assure database port
    error: Misconfiguration or omission is fatal
    tags: PORT
    type: toml
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    toml_filepattern: "^development-\\d+\\.toml$"
    toml_structure: |
      db: {port: 666 }
      ...

  - name: JSON ASSURE - missing database port
    id: 144
    description: Assure database port
    error: Misconfiguration or omission is fatal
    tags: PORT
    type: json
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    json_filepattern: "^development-\\d+\\.json$"
    json_structure: |
      db: {port: 666 }
      ...

```



### SARIF Output 

```json{8,10,24}
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    ... REDACTED
      "results": [
        {
          "ruleId": "intercept.cc.yml.policy.151: YML ASSURE - INGRESS ENABLED",
          "ruleIndex": 0,
          "level": "note",
          "message": {
            "text": "Assure ingress enabled"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-0677.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "COMPLIANT"
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.151: YML ASSURE - INGRESS ENABLED",
          "ruleIndex": 0,
          "level": "note",
          "message": {
            "text": "Assure ingress enabled"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-2048.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "COMPLIANT"
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.151: YML ASSURE - INGRESS ENABLED",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "Assure ingress enabled"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-4985.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "error validating YAML data against CUE schema: ingress.enabled: conflicting values false and true"
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.152: YML ASSURE - INGRESS DISABLED",
          "ruleIndex": 1,
          "level": "error",
          "message": {
            "text": "Assure ingress disabled"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-0677.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "error validating YAML data against CUE schema: ingress.enabled: conflicting values true and false"
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.152: YML ASSURE - INGRESS DISABLED",
          "ruleIndex": 1,
          "level": "error",
          "message": {
            "text": "Assure ingress disabled"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-2048.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "error validating YAML data against CUE schema: ingress.enabled: conflicting values true and false"
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.152: YML ASSURE - INGRESS DISABLED",
          "ruleIndex": 1,
          "level": "note",
          "message": {
            "text": "Assure ingress disabled"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-4985.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "COMPLIANT"
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.153: YML ASSURE - MEMORY LIMIT",
          "ruleIndex": 2,
          "level": "warning",
          "message": {
            "text": "Assure resources"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-0677.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "Missing required keys "
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.153: YML ASSURE - MEMORY LIMIT",
          "ruleIndex": 2,
          "level": "warning",
          "message": {
            "text": "Assure resources"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-2048.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "Missing required keys "
                  }
                }
              }
            }
          ]
        },
        {
          "ruleId": "intercept.cc.yml.policy.153: YML ASSURE - MEMORY LIMIT",
          "ruleIndex": 2,
          "level": "warning",
          "message": {
            "text": "Assure resources"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/filefilter/development-4985.yml"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "Missing required keys "
                  }
                }
              }
            }
          ]
        }
      ]
    ... REDACTED
```

### Console Output

```sh
│
├ YML Rule # 153
│ Rule name :  YML ASSURE - memory limit
│ Rule description :  Assure resources
│ Impacted Env :  all
│ Confidence :  high
│ Tags :  MEMORY
│ File pattern :  ^development-\d+\.yml$
│ Yml pattern :  logging: { level: "info" }
resources: {limits: { memory: "128Mi" }}
...

│ 
│ Scanning..
│ File : /app/examples/target/filefilter/development-0677.yml
│ Hash : 91a0c5dad1cfa795a8eac22c97edecdf8c32e519f04537f0a25cd2e943145d89
│
 {
-  "logging": {
-    "level": "info"
-  },
   "resources": {
     "limits": {
       "memory": "128Mi"
+      "cpu": "100m"
     }
+    "requests": {
+      "cpu": "100m",
+      "memory": "128Mi"
+    }
   }
+  "affinity": {
+  }
+  "autoscaling": {
+    "enabled": false,
+    "maxReplicas": 100,
+    "minReplicas": 1,
+    "targetCPUUtilizationPercentage": 80
+  }
+  "fullnameOverride": ""
+  "image": {
+    "pullPolicy": "IfNotPresent",
+    "repository": "nginx",
+    "tag": "latest"
+  }
+  "ingress": {
+    "enabled": true,
+    "hosts": [
+      : {
+        "host": "chart-example.local",
+        "paths": [
+          : "/"
+        ]
+      }
+    ]
+  }
+  "nameOverride": ""
+  "nodeSelector": {
+  }
+  "service": {
+    "annotations": {
+    },
+    "port": 88,
+    "type": "ClusterIP"
+  }
+  "tolerations": [
+  ]
 }

│
│ NON COMPLIANT : 
│  Misconfiguration or omission is fatal
│  Missing required keys 

│
├── ─

```


# Run it

```sh
docker pull ghcr.io/xfhg/intercept:latest

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -a examples/policy/assure.yaml

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept assure -t examples/target
```