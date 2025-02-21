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
