FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./
RUN rm -rf internal/web/static
COPY --from=frontend /src/frontend/dist/ ./internal/web/static/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/telecom ./cmd/telecom

FROM alpine:3.22
RUN apk add --no-cache iputils && addgroup -S telecom && adduser -S -G telecom telecom
WORKDIR /app
COPY --from=backend /out/telecom /usr/local/bin/telecom
RUN mkdir -p /data && chown telecom:telecom /data
USER telecom
ENV TELECOM_DATA_DIR=/data TELECOM_PORT=14000
VOLUME ["/data"]
EXPOSE 14000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -q -O - http://127.0.0.1:14000/health || exit 1
ENTRYPOINT ["/usr/local/bin/telecom"]
