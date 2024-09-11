FROM alpine:latest
WORKDIR /usr/local/bin
ARG BINARY
COPY ${BINARY} intercept
RUN chmod +x intercept && apk update && apk add --no-cache ripgrep ca-certificates && rm -rf /var/cache/apk/*
# ENTRYPOINT ["/usr/local/bin/intercept"]
# CMD ["--help"]


