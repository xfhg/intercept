# Policy Example: SCAN-003 Missing HTTP Security Headers

**ID:** SCAN-003
**Name:** Detect missing HTTP security headers
**Description:** This policy ensures that critical HTTP security headers are set in the Nginx configuration. These headers help protect web applications against common attacks like cross-site scripting (XSS), clickjacking, and SSL/TLS stripping.
**Severity:** High (Score: 8)

## Why HTTP Security Headers are Important

HTTP security headers are a fundamental part of web application security. They instruct browsers on how to behave when handling your site's content, mitigating various potential vulnerabilities.

The `SCAN-003` policy checks for the presence of the following headers:

-   **`Strict-Transport-Security` (HSTS):** Forces browsers to interact with your site only via HTTPS, preventing downgrade attacks.
-   **`Content-Security-Policy` (CSP):** Helps prevent XSS attacks by defining which dynamic resources are allowed to load.
-   **`X-Frame-Options`:** Protects against clickjacking by controlling whether your site can be embedded in an iframe.
-   **`X-Content-Type-Options`:** Prevents browsers from MIME-sniffing the content type, which can lead to security risks.
-   **`Referrer-Policy`:** Controls how much referrer information is sent with requests.

## How the Policy Works

This policy is of type `scan`. It uses regular expressions to check each line of the specified `filepattern` (e.g., `nginx.conf`) for the absence of `add_header` directives for the security headers listed above. If any of these headers are not found to be configured, the policy will trigger an error.

**Policy Definition Snippet (from `nginx_insecure.yaml`):**

```yaml
  - id: "SCAN-003 Missing HTTP Security Headers"
    type: "scan"
    filepattern: "nginx.conf"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Detect missing HTTP security headers"
      description: "Ensure that critical HTTP security headers are set in the nginx configuration."
      msg_solution: "Add missing HTTP security headers (Strict-Transport-Security, Content-Security-Policy, etc.)."
      msg_error: "Missing HTTP security headers."
      tags:
        - "security"
        - "headers"
      score: "8"
    _regex:
      - "^(?!.*add_header\s+Strict-Transport-Security)"
      - "^(?!.*add_header\s+Content-Security-Policy)"
      - "^(?!.*add_header\s+X-Frame-Options)"
      - "^(?!.*add_header\s+X-Content-Type-Options)"
      - "^(?!.*add_header\s+Referrer-Policy)"
```

## Example Nginx Configuration

### Failing Example (`playground/targets/scan/nginx_missing_security_headers.conf`)

This configuration will trigger the policy because `Strict-Transport-Security` and `Content-Security-Policy` are missing:

```nginx
server {
    listen 80;
    server_name example.com;

    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";

    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
    }
}
```

### Passing Example (`playground/targets/scan/nginx_with_security_headers.conf`)

This configuration satisfies the policy as all required headers are present:

```nginx
server {
    listen 80;
    server_name example.com;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'";
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";
    add_header Referrer-Policy "strict-origin-when-cross-origin";

    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
    }
}
```

## Remediation

To fix this, ensure that all the listed HTTP security headers are correctly configured in your Nginx `server` block(s) using the `add_header` directive. Refer to Nginx documentation and security best practices for appropriate values for each header.
```
