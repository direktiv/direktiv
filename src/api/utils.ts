import { z } from "zod";

const getAuthHeader = (apiKey: string) => ({
  "direktiv-token": apiKey,
});

/**
 * atm params must alway be defined. I tried to make TS infer the property
 * with
 *
 * type ReturnT<TParams> = {
 *   apiKey: string;
 * } & (TParams extends undefined ? object : { params: Partial<TParams> });
 *
 * but it didn't work. I also tried
 *
 * type ReturnT<TParams> = {
 *   apiKey: string;
 *   params?: TParams;
 * };
 *
 * but this would have the downside that params is always optional. And we would
 * lose typesafety when some api enpoints have required params
 *
 */
type ApiParams<TParams, TPathParams> = {
  apiKey: string;
  params: TParams extends undefined ? undefined : TParams;
  pathParams: TPathParams;
};

export const apiFactory =
  <TSchema, TParams, TPathParams>({
    pathFn: path,
    method,
    schema,
  }: {
    pathFn: (pathParams: TPathParams) => string;
    method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
    schema: z.ZodSchema<TSchema>;
  }): (({
    apiKey,
    params,
    pathParams,
  }: ApiParams<TParams, TPathParams>) => Promise<TSchema>) =>
  async ({ apiKey, params, pathParams }): Promise<TSchema> => {
    const res = await fetch(path(pathParams), {
      method,
      headers: {
        ...(apiKey ? getAuthHeader(apiKey) : {}),
      },
      ...(params ? { body: JSON.stringify(params) } : {}),
    });

    if (res.ok) {
      try {
        return schema.parse(await res.json());
      } catch (error) {
        return Promise.reject(
          `could not format response for ${method} ${path(pathParams)}`
        );
      }
    }
    return Promise.reject(
      `error ${res.status} for ${method} ${path(pathParams)}`
    );
  };
