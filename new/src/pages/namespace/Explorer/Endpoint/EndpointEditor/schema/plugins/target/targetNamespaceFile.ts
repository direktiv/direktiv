import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetNamespaceFileFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetNamespaceFile),
  configuration: z.object({
    namespace: z.string().optional(),
    file: z.string().nonempty(),
    content_type: z.string().optional(),
  }),
});

export type TargetNamespaceFileFormSchemaType = z.infer<
  typeof TargetNamespaceFileFormSchema
>;
