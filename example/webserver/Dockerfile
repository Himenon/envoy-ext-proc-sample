FROM node:14.16.1

WORKDIR /app

COPY package.json server.ts tsconfig.json yarn.lock /app/

RUN yarn install --production

CMD ["yarn", "start"]
