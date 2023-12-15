import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlow),
  configuration: z.object({
    flow: z.string().nonempty(),
    content_type: z.string().nonempty(),
    namespace: z.string().nonempty().optional(),
    async: z.boolean().optional(),
  }),
});
