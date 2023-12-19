import { AclFormSchema } from "./acl";
import { JsInboundFormSchema } from "./jsInbound";
import { RequestConvertFormSchema } from "./requestConvert";
import { z } from "zod";

export const InboundPluginFormSchema = z.discriminatedUnion("type", [
  AclFormSchema,
  JsInboundFormSchema,
  RequestConvertFormSchema,
]);
