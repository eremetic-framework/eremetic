module.exports = {
  ...require('@spotify/web-scripts/config/jest.config.js'),
  testEnvironment: 'jsdom',
  testPathIgnorePatterns: ['/node_modules/', '/build/']
};
