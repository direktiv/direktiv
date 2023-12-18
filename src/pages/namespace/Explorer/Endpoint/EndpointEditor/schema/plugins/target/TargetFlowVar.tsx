import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetFlowVarFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetFlowVar),
  configuration: z.object({
    namespace: z.string().nonempty().optional(),
    flow: z.string().nonempty().optional(),
    variable: z.string().nonempty(),
    content_type: z.string().nonempty(),
  }),
});

export type TargetFlowVarFormSchemaType = z.infer<
  typeof TargetFlowVarFormSchema
>;
