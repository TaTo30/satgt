FROM golang:1.26 AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o satgt .


FROM alpine:3
RUN apk upgrade --no-cache --available \
  && apk add --no-cache chromium-swiftshader chromium-chromedriver 

# This variables must be set by the user when running the container
ENV SATGT_USERNIT=""
ENV SATGT_USERPASSWORD=""

# Default cron variables
ENV CRON_PROCESS_WORKLOADS="*/5 * * * *"

ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/lib/chromium/

# VOLUME /pb_data

EXPOSE 8090

COPY --from=build /app/satgt /usr/local/bin/satgt
ENTRYPOINT ["satgt"]
CMD ["serve", "--http=0.0.0.0:8090"]
