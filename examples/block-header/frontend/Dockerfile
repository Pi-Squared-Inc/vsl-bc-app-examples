FROM oven/bun:alpine AS builder
RUN apk add --no-cache python3 build-base
COPY . /opt/app/web
WORKDIR /opt/app/web
RUN bun install
RUN bun run build

FROM node:lts-alpine

WORKDIR /opt/app

COPY --from=builder /opt/app/web/.next/standalone .
COPY --from=builder /opt/app/web/public ./public
COPY --from=builder /opt/app/web/.next/static ./.next/static

EXPOSE 3000
ENV PORT=3000

ENTRYPOINT [ "node", "server.js" ]
