FROM golang:1-alpine3.15 as build

RUN apk add --update --no-cache \
    ca-certificates

COPY action /tmp/build/action

RUN cd /tmp/build/action && go build -mod=vendor -o tfcloud-provider-push-action

FROM alpine:3

RUN apk add --update --no-cache \
    ca-certificates

RUN adduser -g "TFCloud Provider Push Action Guy" tfcpaguy -D -H -s /bin/false

COPY --from=build  /tmp/build/action/tfcloud-provider-push-action /tfcloud-provider-push-action
COPY               scripts/entrypoint.sh /entrypoint.sh

RUN chown tfcpaguy:tfcpaguy /entrypoint.sh /tfcloud-provider-push-action
RUN chmod 500 /entrypoint.sh /tfcloud-provider-push-action

USER tfcpaguy

ENTRYPOINT /entrypoint.sh