## Build
FROM golang:1.19-alpine AS build

RUN apk add --no-cache ca-certificates openssl
RUN /usr/sbin/update-ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY main.go ./
RUN go mod download

COPY var/ ./var/

#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o ./main
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -a -installsuffix cgo -o ./main

## Deploy
FROM scratch
WORKDIR /

COPY --from=build /app/main /main

ENV TZ=America/New_York
COPY --from=build /app/var/$TZ /usr/share/zoneinfo/$TZ
COPY --from=build /app/var/$TZ /etc/localtime
COPY --from=build /etc/ssl/ /etc/ssl/

EXPOSE 8080

ENTRYPOINT ["/main"]