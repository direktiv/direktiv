export const outboundPluginTypes = {
  jsOutbound: { name: "js-outbound", enterpriseOnly: false },
} as const;

const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;

export const availablePlugins = Object.values(outboundPluginTypes).filter(
  (plugin) => (isEnterprise ? true : plugin.enterpriseOnly === false)
);
