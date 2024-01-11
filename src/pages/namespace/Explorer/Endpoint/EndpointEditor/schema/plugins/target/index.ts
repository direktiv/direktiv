import { filterAvailablePlugins } from "../utils";

export const targetPluginTypes = {
  instantResponse: {
    name: "instant-response",
    enterpriseOnly: false,
  },
  targetFlow: {
    name: "target-flow",
    enterpriseOnly: false,
  },
  targetFlowVar: {
    name: "target-flow-var",
    enterpriseOnly: false,
  },
  targetNamespaceFile: {
    name: "target-namespace-file",
    enterpriseOnly: false,
  },
  targetNamespaceVar: {
    name: "target-namespace-var",
    enterpriseOnly: false,
  },
  targetEvent: {
    name: "target-event",
    enterpriseOnly: true,
  },
} as const;

export const availablePlugins = Object.values(targetPluginTypes).filter(
  filterAvailablePlugins
);
