import { targetPluginTypes } from ".";
import { z } from "zod";

export const InstantResponseFormSchema = z.object({
  type: z.literal(targetPluginTypes.instantResponse),
  configuration: z.object({
    content_type: z.string().nonempty(),
    status_code: z.number().int().positive(),
    status_message: z.string().nonempty(),
  }),
});

export type InstantResponseFormSchemaType = z.infer<
  typeof InstantResponseFormSchema
>;
