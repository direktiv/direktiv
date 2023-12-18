import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetNamespaceFile = z.object({
  type: z.literal(targetPluginTypes.targetNamespaceFile),
  configuration: z.object({
    namespace: z.string().nonempty().optional(),
    file: z.string().nonempty(),
    content_type: z.string().nonempty(),
  }),
});

export type TargetNamespaceFileType = z.infer<typeof TargetNamespaceFile>;
