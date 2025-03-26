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
  obj: Record<string, string>,
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

export const headerDefault: PageElementSchemaType = {
  name: "Header",
  hidden: true,
  content: "This is the header",
  preview: "This is the header",
};

export const footerDefault: PageElementSchemaType = {
  name: "Footer",
  hidden: true,
  content: "This is the footer",
  preview: "This is the footer",
};

export const placeholder1: PageElementSchemaType = {
  name: "Text",
  hidden: false,
  content: "This is a Text...",
  preview: "This is a Text...",
};

export const placeholder2: PageElementSchemaType = {
  name: "Table",
  hidden: false,
  content: [{ header: "Example Header", cell: "unset" }],
  preview: "Placeholder Table",
};

export const placeholder3: PageElementSchemaType = {
  name: "Text",
  hidden: true,
  content: "some more info about...",
  preview: "some more info about...",
};

export const defaultConfig: PageFormSchemaType = {
  header: headerDefault,
  footer: footerDefault,
  layout: [placeholder1, placeholder2, placeholder3],
  direktiv_api: "page/v1",
  path: undefined,
};
