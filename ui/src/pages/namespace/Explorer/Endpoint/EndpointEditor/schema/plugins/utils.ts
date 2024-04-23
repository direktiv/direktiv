import { isEnterprise } from "~/config/env/utils";

type Plugin = { name: string; enterpriseOnly?: boolean };

export const filterAvailablePlugins = (plugin: Plugin) =>
  isEnterprise() ? true : plugin.enterpriseOnly === false;
