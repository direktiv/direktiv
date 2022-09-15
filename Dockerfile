FROM node:18 as build

ARG FULL_VERSION

RUN echo "building $FULL_VERSION"
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY public ./public
COPY src ./src
COPY package.json ./
COPY tsconfig.json ./
COPY yarn.lock ./

RUN yarn install
# If this causes problems on github actions: A potential fix is to change the builder image to `node:alpine`
RUN NODE_OPTIONS=--openssl-legacy-provider REACT_APP_VERSION=$FULL_VERSION yarn build

# production environment
FROM nginx:stable-alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY conf/nginx.conf /etc/nginx/conf.d/default.conf
 
CMD ["nginx", "-g", "daemon off;"]