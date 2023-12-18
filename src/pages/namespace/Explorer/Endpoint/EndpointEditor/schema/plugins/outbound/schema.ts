import { JsOutboundFormSchema } from "./jsOutbound";
import { z } from "zod";

export const OutboundPluginFormSchema = z.discriminatedUnion("type", [
  JsOutboundFormSchema,
]);
