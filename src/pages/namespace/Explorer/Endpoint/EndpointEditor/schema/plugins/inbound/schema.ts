import { AclFormSchema } from "./acl";
import { HeaderManipulationFormSchema } from "./headerManipulation";
import { JsInboundFormSchema } from "./jsInbound";
import { RequestConvertFormSchema } from "./requestConvert";
import { z } from "zod";

export const InboundPluginFormSchema = z.discriminatedUnion("type", [
  AclFormSchema,
  HeaderManipulationFormSchema,
  JsInboundFormSchema,
  RequestConvertFormSchema,
]);
