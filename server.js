"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const next_1 = __importDefault(require("next"));
const dev = process.env.NODE_ENV !== 'production';
const app = (0, next_1.default)({ dev });
const handle = app.getRequestHandler();
(async () => {
    await app.prepare();
    const server = (0, express_1.default)();
    // ヘルスチェック
    server.get('/healthz', (_, res) => res.status(200).send('ok'));
    // それ以外は Next.js に渡す
    server.all('*', (req, res) => handle(req, res));
    const port = parseInt(process.env.PORT || '8080', 10);
    server.listen(port, () => console.log(`> Ready on :${port}`));
})();
