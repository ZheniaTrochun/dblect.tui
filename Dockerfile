# Frontend build stage
FROM node:22-slim AS frontend-builder
WORKDIR /app
COPY web/dblect_tui-web/package*.json ./
RUN npm install
COPY web/dblect_tui-web/ ./
RUN npm run build

WORKDIR /app-be
COPY web/dblect_tui-web-backend/package*.json ./
RUN npm install
COPY web/dblect_tui-web-backend/ ./


FROM golang:1.25.8 AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /dblect .


FROM alpine/git AS content

WORKDIR /content
RUN git clone https://github.com/ZheniaTrochun/db-intro-course.git


FROM node:22-slim

RUN apt-get update && apt-get install -y nginx && rm -rf /var/lib/apt/lists/* && rm -f /etc/nginx/sites-enabled/default

COPY --from=builder /usr/src/app/web/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /dblect /usr/local/bin/
COPY --from=frontend-builder /app/dist/ /app/public/
COPY --from=frontend-builder /app-be/ /app/web-be/
COPY --from=content /content/db-intro-course/lectures/ /lectures/
COPY --from=builder /usr/src/app/start.sh /start.sh
COPY --from=builder /usr/src/app/ssh-keys/ /ssh-keys/

CMD ["/start.sh"]
