FROM node:18-buster as build

ARG FULL_VERSION

RUN echo "building $FULL_VERSION"
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY . .

ENV NODE_ENV=production
RUN npm install
# If this causes problems on github actions: A potential fix is to change the builder image to `node:alpine`
RUN VITE_APP_VERSION=$FULL_VERSION npm run build

# production environment
FROM nginx:stable-alpine

COPY --from=build /app/dist /usr/share/nginx/html
COPY conf/nginx.conf /etc/nginx/conf.d/default.conf

CMD ["nginx", "-g", "daemon off;"]