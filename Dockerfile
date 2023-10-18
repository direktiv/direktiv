FROM node:18.18.1 as build

ARG FULL_VERSION

RUN echo "building $FULL_VERSION"
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY . .

RUN yarn install
# If this causes problems on github actions: A potential fix is to change the builder image to `node:alpine`
RUN VITE_APP_VERSION=$FULL_VERSION yarn build

# production environment
FROM nginx:bookworm

COPY --from=build /app/dist /usr/share/nginx/html
COPY conf/nginx.conf /etc/nginx/conf.d/default.conf

COPY conf/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]