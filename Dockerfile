FROM golang:1.13 as builder

WORKDIR /app
COPY . .

RUN make

FROM alpine as fetcher

WORKDIR /app

RUN apk --update add curl \
 && curl -q -sSL --max-time 10 -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem

FROM scratch

EXPOSE 1323

HEALTHCHECK --retries=10 CMD [ "./istravayou-api", "-url", "http://localhost:1323/health" ]
ENTRYPOINT [ "./istravayou-api" ]

COPY --from=builder /app/bin/istravayou-api /