import { InboundPluginFormSchemaType } from "../../schema/plugins/inbound/schema";
import { JsInboundFormSchemaType } from "../../schema/plugins/inbound/jsInbound";
import { JsOutboundFormSchemaType } from "../../schema/plugins/outbound/jsOutbound";
import { OutboundPluginFormSchemaType } from "../../schema/plugins/outbound/schema";
import { RequestConvertFormSchemaType } from "../../schema/plugins/inbound/requestConvert";
import { inboundPluginTypes } from "../../schema/plugins/inbound";
import { outboundPluginTypes } from "../../schema/plugins/outbound";

export const treatEmptyStringAsUndefined = (value: unknown) => {
  if (value === "") {
    return undefined;
  }
  return value;
};

export const getRequestConvertConfigAtIndex = (
  fields: InboundPluginFormSchemaType[] | undefined,
  index: number | undefined
): RequestConvertFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === inboundPluginTypes.requestConvert
    ? plugin.configuration
    : undefined;
};

export const getJsInboundConfigAtIndex = (
  fields: InboundPluginFormSchemaType[] | undefined,
  index: number | undefined
): JsInboundFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === inboundPluginTypes.jsInbound
    ? plugin.configuration
    : undefined;
};

export const getJsOutboundConfigAtIndex = (
  fields: OutboundPluginFormSchemaType[] | undefined,
  index: number | undefined
): JsOutboundFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === outboundPluginTypes.jsOutbound
    ? plugin.configuration
    : undefined;
};
