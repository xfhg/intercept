
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY release/interceptl /usr/local/bin/intercept
COPY .ignore /usr/local/bin/.ignore
RUN chmod +x /usr/local/bin/intercept
CMD ["intercept"]