import {
  PageElementSchemaType,
  PageFormSchema,
  PageFormSchemaType,
} from "./schema";
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
      content: { type: "Text", content: "Example Text" },
      hidden: true,
    },
  ],
};

export const defaultPageFileYaml = jsonToYaml(defaultPageFileJson);

export const extractKeysFromJSON = (
  obj: Record<string, string>,
  parentKey = ""
): string[] => {
  let keys: string[] = [];

  for (const key in obj) {
    const fullKey = parentKey ? `${parentKey}.${key}` : key;

    if (typeof obj[key] === "object" && obj[key] !== null) {
      keys = keys.concat(extractKeysFromJSON(obj[key], fullKey));
    } else {
      keys.push(fullKey);
    }
  }

  return keys;
};

export const placeholder1: PageElementSchemaType = {
  name: "Text",
  hidden: false,
  content: { type: "Text", content: "This is a Text..." },
  preview: "This is a Text...",
};

export const placeholder2: PageElementSchemaType = {
  name: "Table",
  hidden: false,
  content: {
    type: "Table",
    content: [{ header: "Example Header", cell: "- no data -" }],
  },
  preview: "Placeholder Table",
};

export const placeholder3: PageElementSchemaType = {
  name: "Text",
  hidden: true,
  content: { type: "Text", content: "some more info about..." },
  preview: "some more info about...",
};

export const defaultConfig: PageFormSchemaType = {
  layout: [placeholder1, placeholder2, placeholder3],
  direktiv_api: "page/v1",
  path: undefined,
};
