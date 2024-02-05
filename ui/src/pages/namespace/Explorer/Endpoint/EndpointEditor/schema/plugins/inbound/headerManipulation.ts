import { inboundPluginTypes } from ".";
import { z } from "zod";

export const HeaderManipulationFormSchema = z.object({
  type: z.literal(inboundPluginTypes.headerManipulation.name),
  configuration: z.object({
    headers_to_remove: z.array(z.string()).optional(),
  }),
});

export type HeaderManipulationFormSchemaType = z.infer<
  typeof HeaderManipulationFormSchema
>;
