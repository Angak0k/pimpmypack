FROM alpine:3.22.2

LABEL maintainer="romain@alki.earth"

RUN apk update && \
    apk upgrade --no-cache

COPY pimpmypack /bin/pimpmypack

ENTRYPOINT ["/bin/pimpmypack"]