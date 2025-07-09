// src/server.js
const http = require('http');
const port = process.env.PORT || 8080;

const handler = (req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('👋 Picca is alive!\n');
};

http.createServer(handler).listen(port, () => {
  console.log(`Server listening on port ${port}`);
});
