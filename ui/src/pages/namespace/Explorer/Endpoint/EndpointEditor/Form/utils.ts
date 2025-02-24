import { Control, useWatch } from "react-hook-form";
import {
  EndpointFormSchema,
  EndpointFormSchemaType,
  EndpointsPluginsSchema,
  EndpointsPluginsSchemaType,
  XDirektivConfigSchema,
  XDirektivConfigSchemaType,
} from "../schema";

export const useSortedValues = (
  control: Control<EndpointFormSchemaType>
): EndpointFormSchemaType => {
  const watchedValues = useWatch({
    control,
  });

  const sortedRootLevelFields = EndpointFormSchema.keyof().options;
  const sortedRootLevel = sortedRootLevelFields.reduce<EndpointFormSchemaType>(
    (object, key) => ({ ...object, [key]: watchedValues[key] }),
    {} as EndpointFormSchemaType
  );

  const sortedDirektivConfigFields = XDirektivConfigSchema.keyof().options;
  const sortedDirektivConfig =
    sortedDirektivConfigFields.reduce<XDirektivConfigSchemaType>(
      (object, key) => {
        const configToAdd = watchedValues?.["x-direktiv-config"]?.[key]
          ? { [key]: watchedValues?.["x-direktiv-config"]?.[key] }
          : {};
        return { ...object, ...configToAdd };
      },
      {} as XDirektivConfigSchemaType
    );

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

  return {
    ...sortedRootLevel,
    "x-direktiv-config": {
      ...sortedDirektivConfig,
      plugins: hasPlugins ? sortedPlugins : undefined,
    },
  };
};
