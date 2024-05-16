type Plugin = { name: string; enterpriseOnly?: boolean };

export const isPluginAvailable = (plugin: Plugin, isEnterprise: boolean) =>
  isEnterprise ? true : plugin.enterpriseOnly === false;
