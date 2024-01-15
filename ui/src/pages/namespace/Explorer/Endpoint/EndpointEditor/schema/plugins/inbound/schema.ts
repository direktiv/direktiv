import { AclFormSchema } from "./acl";
import { EventFilterFormSchema } from "./eventFilter";
import { JsInboundFormSchema } from "./jsInbound";
import { RequestConvertFormSchema } from "./requestConvert";
import { z } from "zod";

export const InboundPluginFormSchema = z.discriminatedUnion("type", [
  AclFormSchema,
  JsInboundFormSchema,
  RequestConvertFormSchema,
  EventFilterFormSchema,
]);

export type InboundPluginFormSchemaType = z.infer<
  typeof InboundPluginFormSchema
>;
