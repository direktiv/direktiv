import { Control, useWatch } from "react-hook-form";
import { EndpointFormSchemaType, EndpointsPluginsSchema } from "../schema";

import { routeMethods } from "~/api/gateway/schema";

export const useSortedValues = (control: Control<EndpointFormSchemaType>) => {
  const watchedValues = useWatch({ control });

  const sortedConfig = { ...watchedValues["x-direktiv-config"] };

  if (sortedConfig?.plugins) {
    const pluginFields = EndpointsPluginsSchema.keyof().options;
    pluginFields.forEach((key) => {
      const value = sortedConfig.plugins?.[key];

      if (Array.isArray(value) && value.length && sortedConfig.plugins) {
        delete sortedConfig.plugins[key];
      }
    });
    if (Object.keys(sortedConfig.plugins).length === 0) {
      delete sortedConfig.plugins;
    }
  }

  const methods = Array.from(routeMethods).reduce((acc, method) => {
    const value = watchedValues[method];
    if (value !== undefined) {
      acc[method] = value;
    }
    return acc;
  }, {} as Partial<EndpointFormSchemaType>);

  return {
    "x-direktiv-api": watchedValues["x-direktiv-api"],
    "x-direktiv-config": sortedConfig,
    ...methods,
  };
};

export type errorsType = Record<string, unknown>;

// Below are helper functions to flatten the zod errors into a more usable format that
// can be passed on to the FormErrors component.
// zods flatten() function is not used because it returns an array of errors, which is not
// what we want. We want a single error object with the path and message for each error.

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return (
    typeof value === "object" && value !== null && value.constructor === Object
  );
}
export const flattenErrors = (
  errors: errorsType,
  prefix = "",
  seen = new WeakSet()
): Array<{ path: string; message: string }> => {
  let result: Array<{ path: string; message: string }> = [];
  if (seen.has(errors)) return result;
  seen.add(errors);

  Object.entries(errors).forEach(([key, value]) => {
    const newKey = prefix ? `${prefix}.${key}` : key;
    if (isPlainObject(value)) {
      if ("message" in value && Object.keys(value).length === 1) {
        result.push({
          path: newKey,
          message: (value as { message: string }).message,
        });
      } else {
        if (
          "message" in value &&
          typeof (value as { message: unknown }).message === "string"
        ) {
          result.push({
            path: newKey,
            message: (value as { message: string }).message,
          });
        }
        result = result.concat(
          flattenErrors(value as errorsType, newKey, seen)
        );
      }
    }
  });

  return result;
};
