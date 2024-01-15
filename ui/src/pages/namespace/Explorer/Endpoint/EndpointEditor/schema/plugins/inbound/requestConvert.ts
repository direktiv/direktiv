import { inboundPluginTypes } from ".";
import { z } from "zod";

export const RequestConvertFormSchema = z.object({
  type: z.literal(inboundPluginTypes.requestConvert.name),
  configuration: z.object({
    omit_headers: z.boolean(),
    omit_queries: z.boolean(),
    omit_body: z.boolean(),
    omit_consumer: z.boolean(),
  }),
});

export type RequestConvertFormSchemaType = z.infer<
  typeof RequestConvertFormSchema
>;
