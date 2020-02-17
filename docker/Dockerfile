# build stage
FROM golang:alpine AS build-env
WORKDIR /app
RUN apk --no-cache add build-base git gcc
ADD . /app
RUN go get -d . && go build main.go

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /app/main /app/
EXPOSE 8000

ENTRYPOINT ./main
