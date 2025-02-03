import { BaseFileFormSchema, BaseFileFormSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { ZodError } from "zod";

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

const defaultBaseFileJson: BaseFileFormSchemaType = {
  data: {
    spec: {
      openapi: "3.0.0",
      info: {
        title: "Default Title",
        version: "1.0.0",
        description: "Default description",
      },
      paths: {},
    },
    file_path: "",
    errors: [],
  },
};

export const defaultBaseFileYaml = jsonToYaml(defaultBaseFileJson);
