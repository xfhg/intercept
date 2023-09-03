
FROM alpine:latest
COPY bin/interceptl /app/intercept
COPY .ignore /app/.ignore
RUN chmod +x /app/intercept 
CMD ["intercept"]