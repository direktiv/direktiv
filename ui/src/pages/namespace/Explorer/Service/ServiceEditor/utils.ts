import { ServiceFormSchema, ServiceFormSchemaType } from "./schema";

import { ZodError } from "zod";

export const omitEmptyFields = (obj: Record<string, unknown>) =>
  Object.fromEntries(
    Object.entries(obj).filter(([_, value]) => {
      // omit empty string
      if (value === "") {
        return false;
      }
      // omit empty array
      if (Array.isArray(value) && value.length === 0) {
        return false;
      }
      return true;
    })
  );

type SerializeReturnType =
  | [ServiceFormSchemaType, undefined]
  | [undefined, ZodError<ServiceFormSchemaType>];

export const serializeServiceFile = (json: string): SerializeReturnType => {
  let parsedJson: unknown;
  try {
    parsedJson = JSON.parse(json);
  } catch (e) {
    parsedJson = null;
  }

  const jsonParsed = ServiceFormSchema.safeParse(parsedJson);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};

export const defaultServiceFileJson: ServiceFormSchemaType = {
  image: "ealen/echo-server:latest",
};
