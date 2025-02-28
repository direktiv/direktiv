import { EndpointFormSchema, EndpointFormSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { DeepPartialSkipArrayKey } from "react-hook-form";
import { ZodError } from "zod";

type SerializeReturnType =
  | [DeepPartialSkipArrayKey<EndpointFormSchemaType>, undefined]
  | [undefined, ZodError<EndpointFormSchemaType>];

export const serializeEndpointFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);

  const jsonParsed = EndpointFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [normalizeEndpointObject(jsonParsed.data), undefined];
  }

  return [undefined, jsonParsed.error];
};

const defaultEndpointFileJson: EndpointFormSchemaType = {
  "x-direktiv-api": "endpoint/v2",
};

/**
 * this fucntion parses the endpoint config and sorts all the keys recursively. However,
 * it will make sure that all keys starting with x-direktiv- will always be at the top.
 */
export const normalizeEndpointObject = (
  data: DeepPartialSkipArrayKey<EndpointFormSchemaType>
) => {
  if (!data) {
    return data;
  }

  return deepSortObject(data, (a, b) => {
    if (a.startsWith("x-direktiv") && !b.startsWith("x-direktiv")) {
      return -1;
    }
    if (b.startsWith("x-direktiv") && !a.startsWith("x-direktiv")) {
      return 1;
    }
    return a.localeCompare(b);
  });
};

export const defaultEndpointFileYaml = jsonToYaml(defaultEndpointFileJson);

export const deepSortObject = <T extends object>(
  obj: T,
  compare?: (a: string, b: string) => number
): T => {
  if (typeof obj !== "object" || obj === null) {
    return obj; // Return primitives and null as is
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => deepSortObject(item, compare)) as T; // recursively sort array values
  }

  const sortedKeys = Object.keys(obj).sort(compare);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const sortedObj: { [key: string]: any } = {}; // Use index signature

  for (const key of sortedKeys) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sortedObj[key] = deepSortObject((obj as any)[key], compare); // recursively sort object values
  }

  return sortedObj as T;
};
