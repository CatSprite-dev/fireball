FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install 
COPY frontend ./
RUN npm run build

FROM alpine:3.21
WORKDIR /srv
COPY --from=frontend-builder /app/frontend/dist ./
CMD ["sh"]
