import express from 'express';
import next from 'next';

const dev  = process.env.NODE_ENV !== 'production';
const app  = next({ dev });
const handle = app.getRequestHandler();

(async () => {
  await app.prepare();
  const server = express();

  // ヘルスチェック
  server.get('/healthz', (_, res) => res.status(200).send('ok'));

  // それ以外は Next.js に渡す
  server.all('*', (req, res) => handle(req, res));

  const port = parseInt(process.env.PORT || '8080', 10);
  server.listen(port, () => console.log(`> Ready on :${port}`));
})();
