import { isEnterprise } from "~/config/env/utils";
import { isPluginAvailable } from "../utils";

export const outboundPluginTypes = {
  jsOutbound: { name: "js-outbound", enterpriseOnly: false },
} as const;

export const useAvailablePlugins = () =>
  Object.values(outboundPluginTypes).filter((plugin) =>
    isPluginAvailable(plugin, isEnterprise())
  );
