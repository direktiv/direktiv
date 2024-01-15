import { AclFormSchemaType } from "../../schema/plugins/inbound/acl";
import { AuthPluginFormSchemaType } from "../../schema/plugins/auth/schema";
import { BasicAuthFormSchemaType } from "../../schema/plugins/auth/basicAuth";
import { GithubWebhookAuthFormSchemaType } from "../../schema/plugins/auth/githubWebhookAuth";
import { InboundPluginFormSchemaType } from "../../schema/plugins/inbound/schema";
import { JsInboundFormSchemaType } from "../../schema/plugins/inbound/jsInbound";
import { JsOutboundFormSchemaType } from "../../schema/plugins/outbound/jsOutbound";
import { KeyAuthFormSchemaType } from "../../schema/plugins/auth/keyAuth";
import { OutboundPluginFormSchemaType } from "../../schema/plugins/outbound/schema";
import { RequestConvertFormSchemaType } from "../../schema/plugins/inbound/requestConvert";
import { authPluginTypes } from "../../schema/plugins/auth";
import { inboundPluginTypes } from "../../schema/plugins/inbound";
import { outboundPluginTypes } from "../../schema/plugins/outbound";

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

export const getAclConfigAtIndex = (
  fields: InboundPluginFormSchemaType[] | undefined,
  index: number | undefined
): AclFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === inboundPluginTypes.acl
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

export const getBasicAuthConfigAtIndex = (
  fields: AuthPluginFormSchemaType[] | undefined,
  index: number | undefined
): BasicAuthFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === authPluginTypes.basicAuth
    ? plugin.configuration
    : undefined;
};

export const getKeyAuthConfigAtIndex = (
  fields: AuthPluginFormSchemaType[] | undefined,
  index: number | undefined
): KeyAuthFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === authPluginTypes.keyAuth
    ? plugin.configuration
    : undefined;
};

export const getGithubWebhookAuthConfigAtIndex = (
  fields: AuthPluginFormSchemaType[] | undefined,
  index: number | undefined
): GithubWebhookAuthFormSchemaType["configuration"] | undefined => {
  const plugin = index !== undefined ? fields?.[index] : undefined;
  return plugin?.type === authPluginTypes.githubWebhookAuth
    ? plugin.configuration
    : undefined;
};
