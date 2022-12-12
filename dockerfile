FROM alpine

FROM scratch
COPY --from=0  /etc/ssl/certs  /etc/ssl/certs
COPY ./dist /.
COPY ./build_tag.txt /
EXPOSE 80

CMD ["./kubescape-config-service"]