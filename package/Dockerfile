FROM alpine as stage
RUN addgroup -S -g 10005 dayzsa && adduser -S -u 10005 -G dayzsa dayzsa
RUN apk update && apk add --no-cache ca-certificates
RUN \
    mkdir -p /etc/dayzsa && \
    chown -R dayzsa:dayzsa /etc/dayzsa && \
    chmod 750 /etc/dayzsa
COPY package/config.container.yaml /etc/dayzsa/config.yaml

FROM scratch
COPY --from=stage /etc/passwd /etc/passwd
COPY --from=stage /etc/group /etc/group
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=stage /etc/dayzsa /etc/dayzsa
COPY dayzsa-exporter /dayzsa-exporter
USER dayzsa
ENTRYPOINT ["/dayzsa-exporter"]
