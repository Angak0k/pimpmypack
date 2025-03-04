FROM alpine:3.21.3

LABEL maintainer="romain@alki.earth"

RUN apk update && \
    apk upgrade --no-cache

COPY pimpmypack /bin/pimpmypack

ENTRYPOINT ["/bin/pimpmypack"]