import { z } from "zod";

/**
 * ErrorJson is the response body format used when the backend returns an error.
 */
const ErrorJson = z
  .object({
    code: z.string().or(z.number()).optional(),
    message: z.string().optional(),
  })
  .passthrough()
  .optional();

/**
 * ApiErrorSchema is our standardized error schema used to represent api response errors
 * throughout this app. Also see the guard function isApiErrorSchema below.
 */
export const ApiErrorSchema = z.object({
  status: z.number(),
  body: ErrorJson,
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
  let body: ApiErrorSchemaType["body"];
  try {
    body = await getErrorJson(res);
  } catch (error) {
    process.env.NODE_ENV !== "test" && console.error(error);
  }

  return {
    status: res.status,
    body,
  };
};

/**
 * Use isApiErrorSchema() as a guard to check if an error conforms to the ApiErrorSchema
 * (rather than, for example, a standard JavaScript Error originating elsewhere).
 */
export const isApiErrorSchema = (error: unknown): error is ApiErrorSchemaType =>
  ApiErrorSchema.safeParse(error).success;

/**
 * Use getMessageFromApiError(error) to extract the human readable error message.
 * @param error error with unknown type
 * @returns message or undefined if the error has an incompatible format.
 */
export const getMessageFromApiError = (error: unknown) =>
  isApiErrorSchema(error) ? error.body?.message : undefined;

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
    if (error.status === 403 || error.status === 401) {
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

/**
 * Used to type useQuery methods, which may be either an ApiError or other JavaScript error
 * in case something else goes wrong.
 */
export type QueryErrorType = ApiErrorSchemaType | Error;
