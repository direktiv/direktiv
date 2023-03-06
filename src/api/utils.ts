import { z } from "zod";

export const getApiHeaders = (apiKey: string) => ({
  "direktiv-token": apiKey,
});

export function apiFactory<T>({
  path,
  method,
  schema,
}: {
  path: string;
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  schema: z.ZodSchema<T>;
}): ({ apiKey }: { apiKey: string }) => Promise<T> {
  return async ({ apiKey }: { apiKey: string }): Promise<T> => {
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
}
