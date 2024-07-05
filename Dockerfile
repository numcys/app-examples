FROM node:14-alpine

WORKDIR /app

COPY . .

EXPOSE 3000

CMD ["node", "app.js"]
