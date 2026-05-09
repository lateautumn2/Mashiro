FROM node:20-alpine AS builder
WORKDIR /app

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN npm i -g pnpm && pnpm install --frozen-lockfile

COPY frontend/ ./
RUN pnpm build

FROM nginx:1.27-alpine
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
