const next = require('next');
const dev = process.env.NODE_ENV !== 'production';
const app = next({ dev, dir: './src' });

app.prepare().then(() => {
  const path = `${app.nextConfig.distDir}/server/${app.nextConfig.serverBundleName}`;
  return require(path).default(app)
});

export {}
