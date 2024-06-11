FROM --platform=$BUILDPLATFORM node:18.18.1 as build
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY yarn.lock .
COPY package.json .
RUN yarn install

COPY assets assets
COPY conf conf
COPY public public
COPY src src

COPY test test
COPY .env.example .
COPY .eslintrc.js .
COPY .nvmrc .
COPY index.html .
COPY postcss.config.cjs .
COPY tailwind.config.cjs .
COPY tsconfig.json .
COPY vite.config.ts .

RUN yarn build

# production environment
FROM nginx:bookworm

COPY --from=build /app/dist /usr/share/nginx/html
COPY conf/nginx.conf /etc/nginx/conf.d/default.conf

COPY conf/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]