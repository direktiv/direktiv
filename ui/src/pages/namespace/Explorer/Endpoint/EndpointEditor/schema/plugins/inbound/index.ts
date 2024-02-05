import { filterAvailablePlugins } from "../utils";

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
  headerManipulation: {
    name: "header-manipulation",
    enterpriseOnly: true,
  },
  eventFilter: {
    name: "event-filter",
    enterpriseOnly: true,
  },
} as const;

export const availablePlugins = Object.values(inboundPluginTypes).filter(
  filterAvailablePlugins
);
