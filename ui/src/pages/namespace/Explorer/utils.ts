import { parse, stringify } from "yaml";

/**
 * a wrapper around the stringify method of yaml but it will serialize an
 * empty object to an empty string instead of "{}". If an object has multiple
 * keys, that are all set to undefined, it will be considered empty as well.
 */
export const jsonToYaml = (json: Record<string, unknown>) => {
  const isEmptyObject = Object.values(json).filter(Boolean).length === 0;
  return isEmptyObject ? "" : stringify(json);
};

export const yamlToJsonOrNull = (yaml: string) => {
  let json;
  try {
    json = parse(yaml);
  } catch (e) {
    json = null;
  }
  return json;
};

export const treatEmptyStringAsUndefined = (value: unknown) => {
  if (value === "") {
    return undefined;
  }
  return value;
};

export const treatAsNumberOrUndefined = (value: unknown) => {
  const parsed = parseInt(`${value}`, 10);
  if (isNaN(parsed)) {
    return undefined;
  }
  return parsed;
};
