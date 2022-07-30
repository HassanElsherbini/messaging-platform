const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = (app) => {
  console.log("PROXY", process.env.REACT_APP_PROXY_HOST);
  app.use(
    "/api",
    createProxyMiddleware({
      target: process.env.REACT_APP_PROXY_HOST,
      changeOrigin: true,
    })
  );
};
