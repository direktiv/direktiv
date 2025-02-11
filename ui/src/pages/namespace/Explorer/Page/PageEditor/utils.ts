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
    {
      name: "Text",
      preview: "Example Text",
      content: "Example Text",
      hidden: true,
    },
  ],
};

export const defaultPageFileYaml = jsonToYaml(defaultPageFileJson);

export type KeyWithDepth = {
  key: string;
  depth: number;
};

export const extractKeysWithDepth = (
  obj: Record<string, any>,
  depth = 0,
  parentKey = ""
): KeyWithDepth[] => {
  let keysWithDepth: KeyWithDepth[] = [];

  for (const key in obj) {
    const fullKey = parentKey ? `${parentKey}.${key}` : key;

    if (typeof obj[key] === "object" && obj[key] !== null) {
      keysWithDepth = keysWithDepth.concat(
        extractKeysWithDepth(obj[key], depth + 1, fullKey)
      );
    } else {
      keysWithDepth.push({ key: fullKey, depth });
    }
  }

  return keysWithDepth;
};
