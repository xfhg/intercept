# RUNTIME Policies

RUNTIME policies serve as dynamic, real-time validation tools for your server configurations and system states. Unlike static code analysis or pre-deployment checks, RUNTIME policies actively assess the live environment, ensuring continuous compliance and optimal system health.
These policies leverage the power of tools like Goss and Serverspec to provide comprehensive server validation:

### Real-time Verification: 
RUNTIME policies continuously monitor and verify server configurations, services, and system states in production environments.

### Comprehensive Resource Coverage: 
These policies can validate a wide array of system resources, including:

- Packages and software versions
- File existence, permissions, and contents
- Network configurations (ports, addresses, interfaces)
- Running services and processes
- User and group management
- Kernel parameters
- Mount points
- HTTP endpoints and responses


### Infrastructure as Code Validation: 
RUNTIME policies help ensure that your infrastructure matches its intended state as defined in your Infrastructure as Code (IaC) implementations.

### Agentless Operation: 
These policies can be executed without requiring additional software installation on target servers, minimizing overhead and simplifying deployment.

### Realtime Monitoring: 
These policies can serve as health endpoints, integrating with our own **INTERCEPT OBSERVE** monitoring system to provide real-time insights into system compliance and performance.

By implementing RUNTIME policies, organizations can maintain a proactive stance on system compliance, quickly identify configuration drift, and ensure that production environments consistently meet required standards and specifications. This approach bridges the gap between intended configurations and actual system states, fostering a more reliable, secure, and compliant infrastructure.


## Examples

```yaml{3,16-17,20,33-34}
Policies:
  - id: "RUNTIM-001"
    type: "runtime"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
    metadata:
      name: "RUNTIME System Compliance"
      description: "Enforce system compliance at runtime"
      tags:
        - "config"
        - "runtime"
        - "enforce"
      score: "8"
    _runtime:
      config: runtime/simple_https.yaml

  - id: "RUNTIM-002"
    type: "runtime"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
    metadata:
      name: "RUNTIME System Compliance"
      description: "Enforce system compliance at runtime"
      tags:
        - "config"
        - "runtime"
        - "enforce"
      score: "8"
    _runtime:
      config: runtime/irt_sudoers.yaml
```

### runtime/irt_sudoers.yaml
```yaml
user:
  root:
    exists: true
    uid: 0
    gid: 0
    home: /root
    shell: /bin/bash
    groups:
      - root
  authorized_user:
    exists: true
    uid: 1001
    gid: 1001
    home: /home/authorized_user
    shell: /bin/bash
    groups:
      - sudo
  unauthorized_user:
    exists: false

file:
  /etc/sudoers:
    exists: true
    mode: "0440"
    owner: root
    group: root
    contents:
      - "%sudo ALL=(ALL) ALL"
      - "!NOPASSWD:ALL"
      - pattern: "ALL=(ALL) NOPASSWD:ALL"
        invert: true
      - pattern: "some_dangerous_entry"
        invert: true

# Check specific user entries in the sudoers file
command:
  "grep '^authorized_user' /etc/sudoers":
    exit-status: 1
  "grep '^unauthorized_user' /etc/sudoers":
    exit-status: 1
```

### runtime/simple_https.yaml

```yaml
http:
    https://www.google.com:
        status: 200
        allow-insecure: false
        no-follow-redirects: false
        timeout: 5000
        body: []
```

::: tip
RUNTIME policy **config** is written as **gossfile** 

https://goss.readthedocs.io/en/stable/gossfile/
:::

