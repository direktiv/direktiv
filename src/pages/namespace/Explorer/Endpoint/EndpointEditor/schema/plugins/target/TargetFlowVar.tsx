import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowVar = z.object({
  type: z.literal(targetPluginTypes.targetFlowVar),
  configuration: z.object({
    namespace: z.string().nonempty().optional(),
    flow: z.string().nonempty().optional(),
    variable: z.string().nonempty(),
    content_type: z.string().nonempty(),
  }),
});

export type TargetFlowVarType = z.infer<typeof TargetFlowVar>;
