FROM node:latest

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production
COPY server.js ./
ADD dist ./dist

ENV PORT="80"

EXPOSE 80
CMD [ "node", "server.js" ]
