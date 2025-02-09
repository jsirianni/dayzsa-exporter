FROM alpine as stage
RUN addgroup -S -g 10005 dayzsa && adduser -S -u 10005 -G dayzsa dayzsa
RUN apk update && apk add --no-cache ca-certificates

FROM scratch
COPY --from=stage /etc/passwd /etc/passwd
COPY --from=stage /etc/group /etc/group
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY dayzsa-exporter /dayzsa-exporter
USER dayzsa
ENTRYPOINT ["/dayzsa-exporter"]
