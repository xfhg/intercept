
FROM alpine:latest
COPY bin/interceptl /usr/local/bin/intercept
RUN chmod +x /usr/local/bin/intercept 
CMD ["intercept"]