---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "INTERCEPT"
  text: "Compliance as Code"
  tagline: The DevSecOps Compliance Toolkit 
  image:
    src: /interceptvEE.png
    alt: INTERCEPT
  actions:
    - theme: brand
      text: Getting Started
      link: /docs/architecture
    - theme: alt
      text: Policy Documentation
      link: /docs/architecture

features:
  - title: ELEGANT
    details: Multiplatform, Subsecond CI, Low footprint, SARIF output, Webhook Integration
    icon: 🛸
  - title: CODE
    details: REGEX patterns, REGO Policies, CUE Lang Schemas, SERVERSPEC Monitoring
    icon: 🧬
  - title: STANDARD
    details: Policy as Code, No custom languages, Reduced complexity, Industry Standards
    icon: 📡


---

##

**INTERCEPT**<Badge type="warning" text="1.0.X" /> is a specialised **DevSecOps toolkit** that provides comprehensive **Application/System/Service/Endpoint** Code Compliance, Security Testing and Monitoring capabilities. It's engineered to help development and security teams swiftly identify and mitigate vulnerabilities/misconfigurations/leaks in both code, API endpoints, running services or system configurations. It is an essential tool for all stages of your software development lifecycle.


### Full software lifecycle compliance in a heartbeat:

```sh
git clone https://github.com/xfhg/intercept.git
docker pull ghcr.io/xfhg/intercept:latest
```

