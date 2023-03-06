import { z } from "zod";

export const getApiHeaders = (apiKey: string) => ({
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
type ApiParams<TParams> = {
  apiKey: string;
  params: TParams extends undefined ? undefined : TParams;
};

export const apiFactory =
  <TSchema, TParams>({
    path,
    method,
    schema,
  }: {
    path: string;
    method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
    schema: z.ZodSchema<TSchema>;
  }): (({ apiKey, params }: ApiParams<TParams>) => Promise<TSchema>) =>
  async ({ apiKey, params }): Promise<TSchema> => {
    const res = await fetch(path, {
      method,
      headers: {
        "direktiv-token": `${apiKey}`,
      },
      ...(params ? { body: JSON.stringify(params) } : {}),
    });

    if (res.ok) {
      try {
        return schema.parse(await res.json());
      } catch (error) {
        return Promise.reject(
          `could not format response for ${method} ${path}`
        );
      }
    }
    return Promise.reject(`error ${res.status} for ${method} ${path}`);
  };
