export const targetPluginTypes = {
  instantResponse: "instant-response",
  targetFlow: "target-flow",
  targetFlowVar: "target-flow-var",
  targetNamespaceFile: "target-namespace-file",
  targetNamespaceVar: "target-namespace-var",
  targetEvent: "target-event",
} as const;

const allPlugins = Object.values(targetPluginTypes);
const enterpriseOnlyPlugins = [targetPluginTypes.targetEvent] as string[];

const enterprisePlugins = allPlugins;
const openSourcePlugins = allPlugins.filter(
  (plugin) => !enterpriseOnlyPlugins.includes(plugin)
);
const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;
export const availablePlugins = isEnterprise
  ? enterprisePlugins
  : openSourcePlugins;
