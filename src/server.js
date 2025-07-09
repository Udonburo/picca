const http = require('http');
const port = process.env.PORT || 8080;

const handler = (req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('\uD83D\uDC4B Picca is alive!\n');
};

const server = http.createServer(handler);

server.listen(port, () => {
  console.log(`Server listening on port ${port}`);
});
