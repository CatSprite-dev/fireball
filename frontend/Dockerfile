FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci 
COPY frontend ./
RUN npm run build

FROM alpine:3.21
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
CMD ["sh", "-c", "rm -rf /srv/* && cp -r /app/frontend/dist/. /srv"]