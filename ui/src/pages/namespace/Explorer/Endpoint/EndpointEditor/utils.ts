import { EndpointFormSchema, EndpointFormSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { ZodError } from "zod";

type SerializeReturnType =
  | [EndpointFormSchemaType, undefined]
  | [undefined, ZodError<EndpointFormSchemaType>];

export const serializeEndpointFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);

  const jsonParsed = EndpointFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [deepSortObject(jsonParsed.data), undefined];
  }

  return [undefined, jsonParsed.error];
};

const defaultEndpointFileJson: EndpointFormSchemaType = {
  "x-direktiv-api": "endpoint/v2",
};

export const defaultEndpointFileYaml = jsonToYaml(defaultEndpointFileJson);

export const deepSortObject = <T extends object>(obj: T): T => {
  if (typeof obj !== "object" || obj === null) {
    return obj; // Return primitives and null as is
  }

  if (Array.isArray(obj)) {
    return obj.map(deepSortObject) as T; // recursively sort array values
  }

  const sortedKeys = Object.keys(obj).sort();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const sortedObj: { [key: string]: any } = {}; // Use index signature

  for (const key of sortedKeys) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sortedObj[key] = deepSortObject((obj as any)[key]); // recursively sort object values
  }

  return sortedObj as T;
};
