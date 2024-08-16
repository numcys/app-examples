const http = require('http');

const port = process.env.PORT

const server = http.createServer((req, res) => {
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Hello World!\nI\'m: ' + process.env.APP + '\nBuild:' + process.env.BUILD + '\nRelease: ' + process.env.RELEASE + '\nrunning on: ' + process.env.RACK);
  console.log('PING');
});

server.listen(port, () => {
  console.log(`Server running at ${port}/`);
});
