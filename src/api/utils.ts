import { z } from "zod";

const getAuthHeader = (apiKey: string) => ({
  "direktiv-token": apiKey,
});

type HTTPMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

export type ResponseParser = <TUrlParams, TSchema>({
  res,
  schema,
  method,
  urlParams,
  path,
}: {
  res: Response;
  schema: z.ZodSchema<TSchema>;
  method: HTTPMethod;
  urlParams: TUrlParams;
  path: (urlParams: TUrlParams) => string;
}) => Promise<TSchema>;

type FactoryParams<TUrlParams, TSchema> = {
  url: (urlParams: TUrlParams) => string;
  method: HTTPMethod;
  schema: z.ZodSchema<TSchema>;
  responseParser?: ResponseParser;
};

/**
 * atm payload and headers must alway be defined. I tried to make TS infer the property
 * with
 *
 * type ReturnT<TPayload> = {
 *   apiKey: string;
 * } & (TPayload extends undefined ? object : { payload: Partial<TPayload> });
 *
 * but it didn't work. I also tried
 *
 * type ReturnT<TPayload> = {
 *   apiKey: string;
 *   payload?: TPayload;
 * };
 *
 * but this would have the downside that payload is always optional. And we would
 * lose typesafety when some api enpoints have a required payload
 *
 */
type ApiParams<TPayload, THeaders, TUrlParams> = {
  apiKey?: string;
  payload: TPayload extends undefined ? undefined : TPayload;
  headers: THeaders extends undefined ? undefined : THeaders;
  urlParams: TUrlParams;
};

type ApiReturnFunction<TPayload, THeaders, TUrlParams, TSchema> = ({
  apiKey,
  payload,
  urlParams,
}: ApiParams<TPayload, THeaders, TUrlParams>) => Promise<TSchema>;

const defaultResponseParser: ResponseParser = async ({
  res,
  schema,
  method,
  urlParams,
  path,
}) => {
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
  try {
    if (parsedResponse) {
      return schema.parse({ ...parsedResponse });
    }
    return schema.parse(null);
  } catch (error) {
    process.env.NODE_ENV !== "test" && console.error(error);
    return Promise.reject(
      `could not format response for ${method} ${path(urlParams)}`
    );
  }
};

/**
 * API Factory
 *
 * @param url the path to the api endpoint
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
  <TSchema, TPayload, THeaders, TUrlParams>({
    url: path,
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
    const res = await fetch(path(urlParams), {
      method,
      headers: {
        ...(headers && typeof headers === "object" ? headers : {}),
        ...(apiKey ? getAuthHeader(apiKey) : {}),
      },
      ...(payload
        ? {
            body:
              typeof payload === "string" ? payload : JSON.stringify(payload),
          }
        : {}),
    });

    if (res.ok) {
      return responseParser({
        res,
        schema,
        method,
        urlParams,
        path,
      });
    }

    try {
      const json = await res.json();
      return Promise.reject(json);
    } catch (error) {
      process.env.NODE_ENV !== "test" && console.error(error);
      return Promise.reject(
        `error ${res.status} for ${method} ${path(urlParams)}`
      );
    }
  };
