const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;
type Plugin = { name: string; enterpriseOnly?: boolean };

export const filterAvailablePlugins = (plugin: Plugin) =>
  isEnterprise ? true : plugin.enterpriseOnly === false;
