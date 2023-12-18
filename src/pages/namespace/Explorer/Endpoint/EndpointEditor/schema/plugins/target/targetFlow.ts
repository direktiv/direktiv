import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlow),
  configuration: z.object({
    namespace: z.string().optional(),
    flow: z.string().nonempty(),
    // technically optional, but we a boolean is hard to represent in a form as not set
    async: z.boolean().default(false),
    content_type: z.string().optional(),
  }),
});

export type TargetFlowFormSchemaType = z.infer<typeof TargetFlowFormSchema>;
