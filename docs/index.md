---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "INTERCEPT"
  text: "Policy as Code Engine"
  tagline: DevSecOps Compliance toolkit
  actions:
    - theme: brand
      text: Getting Started
      link: /docs/quickstart
    - theme: alt
      text: Policy Documentation
      link: /docs/policy-struct

features:
  - title: No daemons, low footprint, SARIF output
    details: 
  - title: REGEX patterns, REGO Policies and CUE Lang Schemas
    details: 
  - title: No custom policy language, reduced complexity
    details:
  - title: Integrated API endpoint compliance checks
    details: 

---

## Getting Started

Get ready for full compliance in a heartbeat.

```sh
git clone https://github.com/xfhg/intercept.git
docker pull ghcr.io/xfhg/intercept:latest
```