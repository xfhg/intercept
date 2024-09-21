# Docker Quickstart 


<br><br>

### 1. Clone the project repository

Clone the repository to set up a test playground and run our examples:

```sh
git clone https://github.com/xfhg/intercept.git
cd intercept
```

### 2. Pull the latest intercept image

Pull the latest image for your platform:

```sh
docker pull ghcr.io/xfhg/intercept:latest-$OS-$ARCH
# Fill accodringly to your platform
docker pull ghcr.io/xfhg/intercept:latest-linux-arm64

```

### 3. From the project root folder

Execute the following command:

```sh
# assuming linux-amd64 platform
docker run \
  --rm \
  -w "$PWD" \
  -v "$PWD":"$PWD" \
  -e TERM=xterm-256color \
  ghcr.io/xfhg/intercept:latest-linux-amd64 \
  intercept audit \
    --policy playground/policies/test_scan.yaml \
    --target playground/targets \
    -vvvv \
    -o _my_first_run
```

::: warning SANDBOX
This document is a work in progress, hang tight.

Follow the examples in the SANDBOX
:::

