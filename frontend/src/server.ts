import express from 'express';
const port = parseInt(process.env.PORT, 10) || 3000;

export default app => {
  const server = express();
  const handle = app.getRequestHandler();

  server.get('/tasks', (req, res) => {
    return app.render(req, res, '/index', req.query);
  });

  server.get('/launch', (req, res) => {
    return app.render(req, res, '/launch', req.query);
  });

  server.get('*', (req, res) => {
    return handle(req, res);
  });

  server.listen(port, err => {
    if (err) throw err;
    console.info(`> Ready on http://localhost:${port}`);
  });
};
