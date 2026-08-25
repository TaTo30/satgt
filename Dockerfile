FROM golang:1.26 AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o satgt .


FROM alpine:3
RUN apk add --no-cache chromium

EXPOSE 8090

# This variables must be set by the user when running the container
ENV SATGT_USERNIT=""
ENV SATGT_PASSWORD=""

# Default cron variables
ENV CRON_PROCESS_WORKLOADS="*/5 * * * *"

COPY --from=build /app/satgt /usr/local/bin/satgt
ENTRYPOINT ["satgt"]
CMD ["serve", "--http=0.0.0.0:8090"]
