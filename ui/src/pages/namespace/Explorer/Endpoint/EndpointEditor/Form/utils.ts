import { Control, useWatch } from "react-hook-form";
import {
  EndpointFormSchema,
  EndpointFormSchemaType,
  EndpointsPluginsSchema,
} from "../schema";

/**
 * The backend will always return the yaml keys in a predefined order, so we
 * need to make sure to match that order when generating a yaml on the client.
 * To have one deterministic order, we use the order of the keys as defined
 * in the schema.
 */
export const useSortedValues = (control: Control<EndpointFormSchemaType>) => {
  const watchedValues = useWatch({
    control,
  });

  const sortedRootLevelFields = EndpointFormSchema.keyof().options;
  const sortedRootLevel = sortedRootLevelFields.reduce(
    (object, key) => ({ ...object, [key]: watchedValues[key] }),
    {}
  );

  const sortedPluginFields = EndpointsPluginsSchema.keyof().options;
  const sortedPlugins = sortedPluginFields.reduce((object, pluginKey) => {
    const pluginToAdd = watchedValues?.plugins?.[pluginKey]
      ? { [pluginKey]: watchedValues?.plugins?.[pluginKey] }
      : {};
    return { ...object, ...pluginToAdd };
  }, {});
  const hasPlugins = Object.keys(sortedPlugins).length > 0;

  return {
    ...sortedRootLevel,
    plugins: hasPlugins ? sortedPlugins : undefined,
  };
};
