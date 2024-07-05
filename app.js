const http = require('http');
const port = "3000";
var express = require('express');
var app = express();

app.get('/check', function (req, res) {
  res.sendStatus(200);
});

app.get('/', function (req, res) {
  res.status(200).send('Hello World!');
});


app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});

