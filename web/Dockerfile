FROM node:22.22.3-alpine AS build

WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM alpine:3.22

WORKDIR /app
COPY --from=build /src/dist /app/dist
VOLUME ["/srv"]
CMD ["sh", "-c", "rm -rf /srv/* && cp -R /app/dist/. /srv/ && tail -f /dev/null"]
