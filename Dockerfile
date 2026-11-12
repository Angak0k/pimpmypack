FROM alpine:3.24.4

LABEL maintainer="romain@alki.earth"

RUN apk update && \
    apk upgrade --no-cache

COPY pimpmypack /bin/pimpmypack

ENTRYPOINT ["/bin/pimpmypack"]