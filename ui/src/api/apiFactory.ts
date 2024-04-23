import { createApiErrorFromResponse } from "./errorHandling";
import { getAuthHeader } from "./utils";
import { z } from "zod";

type HTTPMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

export type ResponseParser = <TSchema>({
  res,
  schema,
}: {
  res: Response;
  schema: z.ZodSchema<TSchema>;
}) => Promise<TSchema>;

type FactoryParams<TUrlParams, TSchema> = {
  url: (urlParams: TUrlParams) => string;
  method: HTTPMethod;
  schema: z.ZodSchema<TSchema>;
  responseParser?: ResponseParser;
};

type ApiParams<TPayload, THeaders, TUrlParams> = {
  apiKey?: string;
  payload?: TPayload;
  headers?: THeaders extends undefined ? undefined : THeaders;
  urlParams: TUrlParams;
};

type ApiReturnFunction<TPayload, THeaders, TUrlParams, TSchema> = ({
  apiKey,
  payload,
  urlParams,
}: ApiParams<TPayload, THeaders, TUrlParams>) => Promise<TSchema>;

/**
 * Pass your own responseParser to apiFactory when needed. This will be invoked
 * in a try/catch block which will log an error if it fails, so no custom
 * error handling is needed here.
 *
 * @param res the response from the fetch api
 * @param schema the schema to parse against at the end
 * @returns a promise, should always return schema.parse(result)
 */
const defaultResponseParser: ResponseParser = async ({ res, schema }) => {
  // if we can not evaluate the response, we have null as the default
  let parsedResponse = null;
  const textResult = await res.text();
  try {
    // try to parse the response as json
    parsedResponse = JSON.parse(textResult);
  } catch (e) {
    // We use the text response under 'body' if its not an empty string
    if (textResult !== "") parsedResponse = { body: textResult };
  }
  if (parsedResponse) {
    return schema.parse(parsedResponse);
  }
  return schema.parse(null);
};

/**
 * API Factory
 *
 * @param url the url to the api endpoint
 * @param method the http method that should be used for the request
 * @param schema the zod schema that the response should be parsed against.
 * This will give us not only the typesafety of the response, it also validates
 * the response at runtime. Runtime validation is important to catch unexpected
 * responses from the api very early in the application lifecycle and give us
 * confidence about the Typescript types. It comes with the downside that the
 * app is more likely to show errors to the user instead of trying to handle
 * them (which does not scale very well when the complexity of an app grows and
 * leads to even worse user experience).
 * @param responseParser A default parser is supplied, but can be overwritten
 * with a custom parser. Creates a zod parsed response based on the fetch API
 * resonse.
 * @returns a Promise that resolves to the zod parsed response.
 */

export const apiFactory =
  <
    TPayload = unknown,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    TSchema = any,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    THeaders = any,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    TUrlParams = any
  >({
    url,
    method,
    schema,
    responseParser = defaultResponseParser,
  }: FactoryParams<TUrlParams, TSchema>): ApiReturnFunction<
    TPayload,
    THeaders,
    TUrlParams,
    TSchema
  > =>
  async ({ apiKey, payload, headers, urlParams }): Promise<TSchema> => {
    const body =
      typeof payload === "string" ? payload : JSON.stringify(payload);

    const res = await fetch(url(urlParams), {
      method,
      headers: {
        ...(headers && typeof headers === "object" ? headers : {}),
        ...(apiKey ? getAuthHeader(apiKey) : {}),
      },
      ...(payload
        ? {
            body,
          }
        : {}),
    });

    if (res.ok) {
      try {
        const result = await responseParser({
          res,
          schema,
        });
        return result;
      } catch (error) {
        process.env.NODE_ENV !== "test" && console.error(error);
        return Promise.reject(
          `could not format response for ${method} ${url(urlParams)}`
        );
      }
    }

    const apiError = await createApiErrorFromResponse(res);
    return Promise.reject(apiError);
  };
