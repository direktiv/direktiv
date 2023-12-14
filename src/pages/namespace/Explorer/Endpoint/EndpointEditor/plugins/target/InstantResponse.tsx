import { targetPluginTypes } from ".";
import { z } from "zod";

export const InstantResposeFormSchema = z.object({
  type: z.literal(targetPluginTypes.instantResponse),
  configuration: z.object({
    content_type: z.string().nonempty(),
    status_code: z.number().int().positive(),
    status_message: z.string().nonempty(),
  }),
});
