import { PageFormSchema, PageFormSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { ZodError } from "zod";

type SerializeReturnType =
  | [PageFormSchemaType, undefined]
  | [undefined, ZodError<PageFormSchemaType>];

export const serializePageFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);

  const jsonParsed = PageFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};

const defaultPageFileJson: PageFormSchemaType = {
  direktiv_api: "page/v1",
  path: "/",
  layout: [
    { name: "header", hidden: true },
    { name: "footer", hidden: true },
  ],
};

export const defaultPageFileYaml = jsonToYaml(defaultPageFileJson);
