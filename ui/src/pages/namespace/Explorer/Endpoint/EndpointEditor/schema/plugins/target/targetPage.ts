import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetPageFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetPage.name),
  configuration: z.object({
    namespace: z.string().optional(),
    flow: z.string().nonempty(),
    async: z.boolean().optional(),
    content_type: z.string().optional(),
  }),
});

export type TargetPageFormSchemaType = z.infer<typeof TargetPageFormSchema>;
