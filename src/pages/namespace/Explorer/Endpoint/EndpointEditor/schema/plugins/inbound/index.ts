export const inboundPluginTypes = {
  acl: "acl",
  jsInbound: "js-inbound",
  requestConvert: "request-convert",
  eventFilter: "event-filter",
} as const;

const allPlugins = Object.values(inboundPluginTypes);
const enterpriseOnlyPlugins = [inboundPluginTypes.eventFilter] as string[];

const enterprisePlugins = allPlugins;
const openSourcePlugins = allPlugins.filter(
  (plugin) => !enterpriseOnlyPlugins.includes(plugin)
);
const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;
export const availablePlugins = isEnterprise
  ? enterprisePlugins
  : openSourcePlugins;
