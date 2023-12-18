import { inboundPluginTypes } from ".";
import { z } from "zod";

export const RequestConvertFormSchema = z.object({
  type: z.literal(inboundPluginTypes.requestConvert),
  configuration: z.object({}),
});

export type RequestConvertFormSchemaType = z.infer<
  typeof RequestConvertFormSchema
>;
