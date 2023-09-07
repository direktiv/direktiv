import { z } from "zod";

export const ApiErrorSchema = z.object({
  response: z.instanceof(Response),
  json: z
    .object({
      code: z.number().optional(),
      message: z.string().optional(),
    })
    .passthrough()
    .optional(),
});

type ApiErrorSchemaType = z.infer<typeof ApiErrorSchema>;

export const createApiErrorFromResponse = async (
  res: Response
): Promise<ApiErrorSchemaType> => {
  let json: ApiErrorSchemaType["json"];
  try {
    json = await res.json();
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

export const getMessageFromApiError = (error: unknown) => {
  if (isApiErrorSchema(error)) {
    return error.json?.message ?? undefined;
  }
  return undefined;
};
