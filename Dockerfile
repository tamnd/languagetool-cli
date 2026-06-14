FROM alpine:3.21

ARG TARGETPLATFORM

RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 10001 languagetool

COPY $TARGETPLATFORM/languagetool /usr/bin/languagetool

USER languagetool

ENTRYPOINT ["/usr/bin/languagetool"]
