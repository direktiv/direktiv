import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetEventFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetEvent.name),
  configuration: z
    .object({
      namespaces: z.array(z.string()),
    })
    .nullable(), // since all fields are optional, we need to make the whole object optional
});

export type TargetEventFormSchemaType = z.infer<typeof TargetEventFormSchema>;
