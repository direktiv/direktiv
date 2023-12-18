import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowVarFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlowVar),
  configuration: z.object({
    namespace: z.string().optional(),
    flow: z.string().nonempty(),
    variable: z.string().nonempty(),
    content_type: z.string().optional(),
  }),
});

export type TargetFlowVarFormSchemaType = z.infer<
  typeof TargetFlowVarFormSchema
>;
