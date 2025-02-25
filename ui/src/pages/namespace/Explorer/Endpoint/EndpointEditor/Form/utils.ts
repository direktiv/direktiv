import { Control, DeepPartialSkipArrayKey, useWatch } from "react-hook-form";
import {
  EndpointFormSchemaType,
  EndpointsPluginsSchema,
  EndpointsPluginsSchemaType,
} from "../schema";

/**
 * useSortedPlugins is a hook that sorts the plugins of the endpoint form data.
 * The sorting is inconsistent after form submission, resulting in the file
 * appearing to have unsaved changes even when it does not.
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
            [key]: watchedValues["x-direktiv-config"].plugins[key],
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
