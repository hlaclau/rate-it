FROM oven/bun:1 AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/.output .output
ENV NODE_ENV=production
ENV PORT=3007
EXPOSE 3007
CMD ["node", ".output/server/index.mjs"]
