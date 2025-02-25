import { Control, DeepPartialSkipArrayKey, useWatch } from "react-hook-form";
import {
  EndpointFormSchemaType,
  EndpointsPluginsSchema,
  EndpointsPluginsSchemaType,
} from "../schema";

/**
 * useSortedPlugins is a hook that sorts the plugins object in
 * the form based on the keys defined in the EndpointsPluginsSchema.
 * It returns a new object with the plugins sorted.
 */
export const useSortedPlugins = (
  control: Control<EndpointFormSchemaType>
): DeepPartialSkipArrayKey<EndpointFormSchemaType> => {
  const watchedValues = useWatch({
    control,
  });

  const sortedPluginsFields = EndpointsPluginsSchema.keyof().options;
  const sortedPlugins = sortedPluginsFields.reduce<EndpointsPluginsSchemaType>(
    (object, key) => {
      const configToAdd = watchedValues?.["x-direktiv-config"]?.plugins?.[key]
        ? {
            [key]: watchedValues?.["x-direktiv-config"]?.plugins?.[key],
          }
        : {};
      return { ...object, ...configToAdd };
    },
    {} as EndpointsPluginsSchemaType
  );

  const hasPlugins = Object.keys(sortedPlugins).length > 0;

  const configWithSortedPlugins =
    hasPlugins && watchedValues?.["x-direktiv-config"]
      ? {
          ["x-direktiv-config"]: {
            ...watchedValues["x-direktiv-config"],
            plugins: sortedPlugins,
          },
        }
      : {};

  return {
    ...watchedValues,
    ...configWithSortedPlugins,
  };
};
