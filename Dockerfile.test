
FROM alpine:latest
COPY bin/interceptl /usr/local/bin/intercept
COPY docker/run.sh /usr/local/bin/run.sh
RUN chmod +x /usr/local/bin/intercept /usr/local/bin/run.sh
CMD ["/usr/local/bin/run.sh"]