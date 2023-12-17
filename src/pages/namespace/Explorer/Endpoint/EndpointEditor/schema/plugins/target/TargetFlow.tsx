import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlow),
  configuration: z.object({
    flow: z.string().nonempty(),
    content_type: z.string().nonempty(),
    namespace: z.string().nonempty().optional(),
    // technically optional, but we a boolean is hard to represent in a form as not set
    async: z.boolean(),
  }),
});

export type TargetFlowFormSchemaType = z.infer<typeof TargetFlowFormSchema>;
