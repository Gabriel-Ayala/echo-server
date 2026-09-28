FROM docker.io/library/golang:1.26.1-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/echo-server

FROM scratch
COPY --from=build /server /server
USER 65532:65532
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/server"]
