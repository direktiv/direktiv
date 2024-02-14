import request from "supertest";

const requestWithHeaders =
  (appConfig, method = "post") =>
  (args) =>
    request(appConfig)[method](args).set("Direktiv-Token", "password");

const customRequest = (appConfig) => ({
  get: requestWithHeaders(appConfig, "get"),
  head: requestWithHeaders(appConfig, "head"),
  post: requestWithHeaders(appConfig, "post"),
  put: requestWithHeaders(appConfig, "put"),
  delete: requestWithHeaders(appConfig, "delete"),
  patch: requestWithHeaders(appConfig, "patch")
});

export default customRequest;