import { Control, useWatch } from "react-hook-form";
import {
  EndpointFormSchema,
  EndpointFormSchemaType,
  EndpointsPluginsSchema,
} from "../schema";

export const useSortedValues = (control: Control<EndpointFormSchemaType>) => {
  const watchedValues = useWatch({ control });

  const configKeys =
    EndpointFormSchema.shape["x-direktiv-config"].keyof().options;

  const sortedConfig = configKeys.reduce(
    (acc, key) => ({
      ...acc,
      [key]: watchedValues?.["x-direktiv-config"]?.[key] ?? [],
    }),
    {} as Record<string, unknown>
  );

  const sortedPluginFields = EndpointsPluginsSchema.keyof().options;
  const sortedPlugins = sortedPluginFields.reduce((acc, pluginKey) => {
    const pluginValue =
      watchedValues?.["x-direktiv-config"]?.plugins?.[pluginKey] ?? [];
    return pluginValue ? { ...acc, [pluginKey]: pluginValue } : acc;
  }, {});

  const hasPlugins = Object.keys(sortedPlugins).length > 0;
  if (hasPlugins) {
    sortedConfig.plugins = sortedPlugins;
  } else {
    delete sortedConfig.plugins;
  }

  return {
    "x-direktiv-api": watchedValues["x-direktiv-api"],
    "x-direktiv-config": sortedConfig,
  };
};
