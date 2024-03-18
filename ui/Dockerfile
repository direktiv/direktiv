FROM node:18.18.1 as build

ARG FULL_VERSION
ARG IS_ENTERPRISE=FALSE
ARG UI_BASE=/

RUN echo "building $FULL_VERSION"
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY yarn.lock yarn.lock
COPY package.json package.json
RUN yarn install

COPY . .

RUN VITE_APP_VERSION=$FULL_VERSION VITE_BASE=${UI_BASE} VITE_IS_ENTERPRISE=${IS_ENTERPRISE} yarn build

# production environment
FROM nginx:bookworm

COPY --from=build /app/dist /usr/share/nginx/html
COPY conf/nginx.conf /etc/nginx/conf.d/default.conf

COPY conf/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]