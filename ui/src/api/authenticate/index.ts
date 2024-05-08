import { ApiErrorSchema } from "../errorHandling";
import { getNamespaces } from "../namespaces/query/get";

export const authenticationKeys = {
  authentication: (apiKey: string | undefined) =>
    [{ scope: "authentication", apiKey }] as const,
};

/**
 * There is no dedicated authentication endpoint. Instead, we query the
 * "version" endpoint. Like with any other endpoint, it will succeed with the
 * correct auth keys and fail otherwise.
 */
export const checkApiKeyAgainstServer = (apiKey?: string) =>
  getNamespaces({
    apiKey,
    urlParams: {},
  })
    .then(() => true)
    /**
     * when the test call fails with a 401 or 403 our auth test fails, any
     * other error should bubble up, like a 500 or schema validation error
     */
    .catch((err) => {
      const parsedError = ApiErrorSchema.safeParse(err);
      if (parsedError.success) {
        const { status } = parsedError.data.response;
        if (status === 401 || status === 403) {
          return false;
        }
      }
      throw err;
    });
