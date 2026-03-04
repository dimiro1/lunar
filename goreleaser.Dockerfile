FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/lunar .

EXPOSE 3000

ENV DATA_DIR=/data \
    EXECUTION_TIMEOUT=300

CMD ["./lunar"]
