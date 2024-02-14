import request from "supertest";

/**
 * checks if the environment variables AUTH_TOKEN_KEY and AUTH_TOKEN_VALUE are set and returns
 * either an array with the key and value or null.
 */
const getAuthHeader = () => {
  if (process.env.AUTH_TOKEN_KEY && process.env.AUTH_TOKEN_VALUE) {
    return [process.env.AUTH_TOKEN_KEY, process.env.AUTH_TOKEN_VALUE];
  }
  return null;
};

const requestWithHeaders =
  (appConfig, method = "post") =>
  (args) => {
    const authHeader = getAuthHeader();

    if (!authHeader) {
      return request(appConfig)[method](args);
    }

    return request(appConfig)
      [method](args)
      .set(...authHeader);
  };

/**
 * overwrites the http methods (get, head, post, put, delete, patch)
 * from supertest and injects a custom header if the environment variables
 * AUTH_TOKEN_KEY and AUTH_TOKEN_VALUE are set.
 */
const customRequest = (appConfig) => ({
  get: requestWithHeaders(appConfig, "get"),
  head: requestWithHeaders(appConfig, "head"),
  post: requestWithHeaders(appConfig, "post"),
  put: requestWithHeaders(appConfig, "put"),
  delete: requestWithHeaders(appConfig, "delete"),
  patch: requestWithHeaders(appConfig, "patch"),
});

export default customRequest;