FROM golang:1.18-alpine3.16 as build

RUN apk add --update --no-cache \
    ca-certificates

COPY action /tmp/build/action

RUN cd /tmp/build/action && go build -mod=vendor -o tfcloud-provider-push-action

FROM alpine:3.21

RUN apk add --update --no-cache \
    ca-certificates \
    dumb-init

RUN adduser -g "TFCloud Provider Push Action Guy" tfcpaguy -D -H -s /bin/false

COPY --from=build  /tmp/build/action/tfcloud-provider-push-action /tfcloud-provider-push-action

RUN chown tfcpaguy:tfcpaguy /tfcloud-provider-push-action
RUN chmod 500 /tfcloud-provider-push-action

USER tfcpaguy

ENTRYPOINT [ "dumb-init", "/tfcloud-provider-push-action" ]