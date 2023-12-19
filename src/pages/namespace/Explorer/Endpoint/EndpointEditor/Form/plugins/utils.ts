import { InboundPluginFormSchemaType } from "../../schema/plugins/inbound/schema";
import { JsInboundFormSchemaType } from "../../schema/plugins/inbound/jsInbound";
import { RequestConvertFormSchemaType } from "../../schema/plugins/inbound/requestConvert";
import { inboundPluginTypes } from "../../schema/plugins/inbound";

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
