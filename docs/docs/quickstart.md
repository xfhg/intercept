# Docker Quickstart 

```sh
git clone https://github.com/xfhg/intercept.git
```

```sh
docker pull ghcr.io/xfhg/intercept:latest
```

from the project root folder:

```sh
docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -a examples/policy/filetype.yaml
```

```sh
docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept yml -t examples/target
```
