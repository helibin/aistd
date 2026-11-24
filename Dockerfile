FROM golang:1.23-alpine AS cli-build
RUN apk add --no-cache make
WORKDIR /src/apps/cli
COPY apps/cli/ ./
RUN sh scripts/release.sh

FROM node:22-alpine AS web-build
WORKDIR /src
COPY package.json package-lock.json ./
COPY apps/web/package.json apps/web/package.json
RUN npm ci
COPY apps/web/ apps/web/
COPY apps/cli/VERSION apps/cli/VERSION
RUN npm run build:web

FROM nginx:1.28-alpine
COPY --from=web-build /src/dist/ /usr/share/nginx/html/
COPY --from=cli-build /src/apps/cli/release/ /usr/share/nginx/html/release/
COPY --from=cli-build /src/apps/cli/release/install.sh /etc/aistd/install.sh
COPY deploy/docker-entrypoint.sh /usr/local/bin/aistd-entrypoint.sh
EXPOSE 80
ENTRYPOINT ["/bin/sh", "/usr/local/bin/aistd-entrypoint.sh"]
