import { ServiceFormSchema, ServiceFormSchemaType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

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

export const serializeServiceFile = (yaml: string): SerializeReturnType => {
  const json = yamlToJsonOrNull(yaml);

  const jsonParsed = ServiceFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return [jsonParsed.data, undefined];
  }

  return [undefined, jsonParsed.error];
};

const defaultServiceFileJson: ServiceFormSchemaType = {
  direktiv_api: "service/v1",
};

export const defaultServiceYaml = jsonToYaml(defaultServiceFileJson);
