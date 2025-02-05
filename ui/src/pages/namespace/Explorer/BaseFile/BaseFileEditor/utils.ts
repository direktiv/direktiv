import { BaseFileFormSchema, BaseFileFormSchemaType } from "./schema";

import { ZodError } from "zod";
import { yamlToJsonOrNull } from "../../utils";

type SerializeReturnType =
  | [BaseFileFormSchemaType, undefined]
  | [undefined, ZodError<BaseFileFormSchemaType>];

export const serializeBaseFileFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);
  const jsonParsed = BaseFileFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};
