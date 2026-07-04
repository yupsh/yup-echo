# Integration harness: build the yup-echo binary for Linux, then run it inside a
# Debian image (GNU coreutils) so its flags, inputs, and outputs can be checked
# against the real `echo` reference. Invoked by `make integration`.
FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/yup-echo .

FROM debian:stable-slim
COPY --from=build /out/yup-echo /usr/local/bin/yup-echo
COPY integration/run.sh /run.sh
ENTRYPOINT ["/bin/sh", "/run.sh"]
