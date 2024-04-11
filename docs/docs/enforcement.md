

# Policy Enforcement Levels

Enforcement levels are a first class concept in allowing pass/fail behavior to be associated separately from the policy logic. This enables any policy to be a warning, allow exceptions, or be absolutely mandatory. These levels can be coupled to environments, different uses of the same policy can have different enforcement levels per environment.

You can set three enforcement levels:

### **Advisory**
- The policy is allowed to fail. However, a warning will be shown to the user or logged.

```yaml
  - fatal: false
  - enforcement: false
  - environment : (all | optional)
  - confidence : low | high
```

### **Soft Mandatory**

- The policy must pass unless an exception is specified. The purpose of this level is to provide a level of privilege separation for a behavior. Additionally, the exception provides non-repudiation since at least the primary actor was explicitly overriding a failed policy.

```yaml{1}
  - fatal: true
  - enforcement: false
  - environment : (all | optional)
  - confidence : low | high
```

### **Hard Mandatory**:
- The policy must pass no matter what. The only way to override a hard mandatory policy is to explicitly remove the policy. It should be used in situations where an exception is not possible.

```yaml{1,2}
  - fatal: true
  - enforcement: true
  - environment : (all | optional)
  - confidence : high
```
