import { AclFormSchema } from "./acl";
import { EventFilterFormSchema } from "./eventFilter";
import { HeaderManipulationFormSchema } from "./headerManipulation";
import { JsInboundFormSchema } from "./jsInbound";
import { RequestConvertFormSchema } from "./requestConvert";
import { z } from "zod";

export const InboundPluginFormSchema = z.discriminatedUnion("type", [
  AclFormSchema,
  JsInboundFormSchema,
  RequestConvertFormSchema,
  EventFilterFormSchema,
  HeaderManipulationFormSchema,
]);

export type InboundPluginFormSchemaType = z.infer<
  typeof InboundPluginFormSchema
>;
