import { z } from "zod";

/**
 * The ApiErrorSchema is a special schema we use to standardize api error handling
 * across the app. It contains the response object from the fetch api, and an
 * optional json object that may contain the error code and message. Since errors
 * are always typed as unknown, we can use the the custom type guard isApiErrorSchema
 * to check if an error conforms to the ApiErrorSchema and have typesafe way to process
 * the error.
 */
const ErrorJson = z
  .object({
    code: z.string().or(z.number()).optional(),
    message: z.string().optional(),
  })
  .passthrough()
  .optional();

export const ApiErrorSchema = z.object({
  response: z.instanceof(Response),
  json: ErrorJson,
});

type ApiErrorSchemaType = z.infer<typeof ApiErrorSchema>;

type ErrorJsonType = z.infer<typeof ErrorJson>;

/**
 * Returns an object describing the error. Works with v1 api as well as v2.
 *
 * Response body format for errors: {
 *   error: {
 *     code: "code",
 *     message: "message",
 *   }
 * }
 */
const getErrorJson = async (res: Response): Promise<ErrorJsonType> => {
  let receivedJson = await res.json();
  receivedJson = receivedJson.error;
  return ErrorJson.parse(receivedJson);
};

export const createApiErrorFromResponse = async (
  res: Response
): Promise<ApiErrorSchemaType> => {
  let json: ApiErrorSchemaType["json"];
  try {
    json = await getErrorJson(res);
  } catch (error) {
    process.env.NODE_ENV !== "test" && console.error(error);
  }

  return {
    response: res,
    json,
  };
};

export const isApiErrorSchema = (error: unknown): error is ApiErrorSchemaType =>
  ApiErrorSchema.safeParse(error).success;

export const getMessageFromApiError = (error: unknown) =>
  isApiErrorSchema(error) ? error.json?.message : undefined;

type PermissionStatus =
  | {
      isAllowed: true;
    }
  | {
      isAllowed: false;
      message?: string;
    };

export const getPermissionStatus = (error: unknown): PermissionStatus => {
  if (isApiErrorSchema(error)) {
    if (error.response.status === 403 || error.response.status === 401) {
      return {
        isAllowed: false,
        message: getMessageFromApiError(error),
      };
    }
  }

  return {
    isAllowed: true,
  };
};
