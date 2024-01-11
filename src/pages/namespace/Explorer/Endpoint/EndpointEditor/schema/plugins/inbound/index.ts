export const inboundPluginTypes = {
  acl: {
    name: "acl",
    enterpriseOnly: false,
  },
  jsInbound: {
    name: "js-inbound",
    enterpriseOnly: false,
  },
  requestConvert: {
    name: "request-convert",
    enterpriseOnly: false,
  },
  eventFilter: {
    name: "event-filter",
    enterpriseOnly: true,
  },
} as const;

const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;

export const availablePlugins = Object.values(inboundPluginTypes).filter(
  (plugin) => (isEnterprise ? true : plugin.enterpriseOnly === false)
);
