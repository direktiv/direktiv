import {
  OpenApiBaseFileFormSchema,
  OpenApiBaseFileFormSchemaType,
} from "./schema";

import { ZodError } from "zod";
import { yamlToJsonOrNull } from "../../utils";

type SerializeReturnType =
  | [OpenApiBaseFileFormSchemaType, undefined]
  | [undefined, ZodError<OpenApiBaseFileFormSchemaType>];

export const serializeBaseFileFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);
  const jsonParsed = OpenApiBaseFileFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};

// I created this function, but ended up not using it. Might as well save it for future us :)
