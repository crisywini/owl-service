FROM golang:1.25-alpine AS build


WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download && go mod verify 

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main.go


#Shell
FROM alpine:3.20 

RUN apk --no-cache add ca-certificates && \
      addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser/
COPY --from=build /app/server .

RUN chown appuser:appgroup ./server

USER appuser

ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["./server"]