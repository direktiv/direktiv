import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetNamespaceVarFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetNamespaceVar),
  configuration: z.object({
    namespace: z.string().nonempty().optional(),
    variable: z.string().nonempty(),
    content_type: z.string().optional(),
  }),
});

export type TargetNamespaceVarFormSchemaType = z.infer<
  typeof TargetNamespaceVarFormSchema
>;
