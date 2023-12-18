import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetNamespaceVar = z.object({
  type: z.literal(targetPluginTypes.targetNamespaceVar),
  configuration: z.object({
    namespace: z.string().nonempty().optional(),
    variable: z.string().nonempty(),
    content_type: z.string().nonempty(),
  }),
});

export type TargetNamespaceVarType = z.infer<typeof TargetNamespaceVar>;
