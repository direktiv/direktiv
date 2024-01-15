import { ApiErrorSchema } from "../errorHandling";
import { getVersion } from "../version/query/get";

export const authenticationKeys = {
  authentication: (apiKey: string | undefined) =>
    [{ scope: "authentication", apiKey }] as const,
};

export const checkApiKeyAgainstServer = (apiKey?: string) =>
  getVersion({
    apiKey,
    urlParams: undefined,
  }) // when test call succeeds and matches the schema our auth test passes
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
