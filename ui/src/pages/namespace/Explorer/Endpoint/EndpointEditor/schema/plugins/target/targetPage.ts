import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetPageFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetPage.name),
  configuration: z.object({
    file: z.string().nonempty(),
  }),
});

export type TargetPageFormSchemaType = z.infer<typeof TargetPageFormSchema>;
