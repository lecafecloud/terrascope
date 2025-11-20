FROM node:22-slim AS builder
WORKDIR /app
COPY package*.json tsconfig.json ./
RUN npm ci
COPY ./ ./
RUN npm run build

FROM node:22-slim AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev

FROM alpine:3.22 AS final
WORKDIR /app
RUN apk add --no-cache nodejs npm && \
    npm install -g serve@14.2.5
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/build/ ./build
RUN addgroup -g 1000 node && \
    adduser -u 1000 -G node -s /bin/sh -D node && \
    chown -R node:node /app
USER node
EXPOSE 3000
ENTRYPOINT ["serve", "-s", "build"]
