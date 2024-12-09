import { isEnterprise } from "~/config/env/utils";
import { isPluginAvailable } from "../utils";

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
    enterpriseOnly: false,
  },
  eventFilter: {
    name: "event-filter",
    enterpriseOnly: true,
  },
} as const;

export const useAvailablePlugins = () =>
  Object.values(inboundPluginTypes).filter((plugin) =>
    isPluginAvailable(plugin, isEnterprise())
  );
