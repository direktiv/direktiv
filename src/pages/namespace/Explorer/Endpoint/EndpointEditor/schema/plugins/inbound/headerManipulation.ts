import { inboundPluginTypes } from ".";
import { z } from "zod";

export const HeaderManipulationFormSchema = z.object({
  type: z.literal(inboundPluginTypes.headerManipulation),
  configuration: z.object({}),
});

export type HeaderManipulationFormSchemaType = z.infer<
  typeof HeaderManipulationFormSchema
>;
