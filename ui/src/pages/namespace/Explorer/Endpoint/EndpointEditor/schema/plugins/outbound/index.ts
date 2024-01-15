import { filterAvailablePlugins } from "../utils";

export const outboundPluginTypes = {
  jsOutbound: { name: "js-outbound", enterpriseOnly: false },
} as const;

export const availablePlugins = Object.values(outboundPluginTypes).filter(
  filterAvailablePlugins
);
