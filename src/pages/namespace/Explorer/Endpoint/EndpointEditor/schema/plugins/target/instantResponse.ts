import { targetPluginTypes } from ".";
import { z } from "zod";

export const InstantResponseFormSchema = z.object({
  type: z.literal(targetPluginTypes.instantResponse),
  configuration: z.object({
    content_type: z.string().optional(),
    status_code: z.number().int().positive(),
    status_message: z.string().optional(),
  }),
});

export type InstantResponseFormSchemaType = z.infer<
  typeof InstantResponseFormSchema
>;
