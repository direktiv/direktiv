FROM --platform=$BUILDPLATFORM node:18.18.1 as build
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY yarn.lock yarn.lock
COPY package.json package.json
RUN yarn install

COPY . .

RUN yarn build

# production environment
FROM --platform=$TARGETPLATFORM nginx:bookworm

COPY --from=build /app/dist /usr/share/nginx/html
COPY conf/nginx.conf /etc/nginx/conf.d/default.conf

COPY conf/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]