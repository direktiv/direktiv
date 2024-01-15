import { isEqual } from "lodash";
import { stringify as jsonToPrettyYamlStringify } from "json-to-pretty-yaml";
import yamljs from "js-yaml";

/**
 * a wrapper around the stringify method of json-to-pretty-yaml
 * but it will serialize an empty object to an empty string instead
 * of "{}"
 */
export const jsonToYaml = (json: Record<string, unknown>) =>
  Object.keys(json).length === 0 ? "" : jsonToPrettyYamlStringify(json);

export const yamlToJsonOrNull = (yaml: string) => {
  let json;
  try {
    json = yamljs.load(yaml);
  } catch (e) {
    json = null;
  }
  return json;
};

/**
 * Compares two yaml strings. Please note that this function will make a
 * structural comparison, so it will return true if the yaml strings are
 * structurally equal, even if the order of the keys is different.
 */
export const compareYamlStructure = (aYaml: string, bYaml: string) =>
  isEqual(yamlToJsonOrNull(aYaml), yamlToJsonOrNull(bYaml));

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
