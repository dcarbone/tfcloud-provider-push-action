FROM golang:1-alpine3.15 as build

RUN apk add --update --no-cache \
    ca-certificates

COPY action /tmp/build/action

RUN cd /tmp/build/action && go build -mod=vendor -o tfcloud-provider-push-action

FROM alpine:3

RUN apk add --update --no-cache \
    ca-certificates

RUN adduser -g "TFCloud Provider Push Action Guy" tfcpaguy -D -H -s /bin/false

COPY --chmod=500 --chown=tfcpaguy --from=build  /tmp/build/action/tfcloud-provider-push-action /tfcloud-provider-push-action
COPY --chmod=500 --chown=tfcpaguy               scripts/entrypoint.sh /entrypoint.sh

USER tfcpaguy

ENTRYPOINT /entrypoint.sh