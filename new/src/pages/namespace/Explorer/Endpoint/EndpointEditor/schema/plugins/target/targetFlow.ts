import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlow.name),
  configuration: z.object({
    namespace: z.string().optional(),
    flow: z.string().nonempty(),
    async: z.boolean().optional(),
    content_type: z.string().optional(),
  }),
});

export type TargetFlowFormSchemaType = z.infer<typeof TargetFlowFormSchema>;
