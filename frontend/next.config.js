const withTypescript = require('@zeit/next-typescript');

const withServerBundle = function withServerBundle(
  entrypoint = './server.js',
  nextConfig = {},
) {
  return {
    ...nextConfig,
    serverBundleName: 'server.js',
    webpack(config, options) {
      // add an entrypoint (typically server.ts) to the bundle (so that babel can run on it)
      const original = config.entry;
      config.entry = () =>
        original().then(entry => {
          if (options.isServer) {
            entry[options.config.serverBundleName] = [entrypoint];
          }
          return entry;
        });

      // change working dir for next-babel-loader
      // so that it reads the babel.config.js in root
      config.module.rules.forEach(rule => {
        if (rule.use.loader === 'next-babel-loader') {
          rule.use.options.cwd = '.';
        }
      });
      if (typeof nextConfig.webpack === 'function') {
        return nextConfig.webpack(config, options);
      }
      return config;
    },
  };
};

module.exports = withTypescript(
  withServerBundle('./server.ts', {
    distDir: '../build',
  }),
);
