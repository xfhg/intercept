# API Policy Examples

This document provides step-by-step example documentation for API policies.

## Example 1: Basic API Policy

Content of `basic_api_policy.yaml`:

```yaml
# Example API Policy
# Define a basic API policy here...
```

Content of `basic_api_target.yaml`:

```yaml
# Example API Target
# Define a basic API target here...
```

To test this policy against its target, run the following command:

```bash
docker run --rm -v $(pwd):/data mypolicytool test /data/basic_api_policy.yaml /data/basic_api_target.yaml
```
