import { z } from "zod";

export const getApiHeaders = (apiKey: string) => ({
  "direktiv-token": apiKey,
});

export const apiFactory =
  <TParams extends object, TSchema>({
    path,
    method,
    schema,
  }: {
    path: string;
    method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
    schema: z.ZodSchema<TSchema>;
  }): (({
    apiKey,
    params,
  }: {
    apiKey: string;
    params: TParams;
  }) => Promise<TSchema>) =>
  async ({ apiKey }: { apiKey: string }): Promise<TSchema> => {
    const res = await fetch(path, {
      method,
      headers: {
        "direktiv-token": `${apiKey}`,
      },
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
