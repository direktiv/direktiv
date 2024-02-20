import { inboundPluginTypes } from ".";
import { z } from "zod";

const HeaderSchema = z.object({
  name: z.string(),
  value: z.string(),
});

const HeaderRemoveSchema = z.object({
  name: z.string(),
});

export const HeaderManipulationFormSchema = z.object({
  type: z.literal(inboundPluginTypes.headerManipulation.name),
  configuration: z.object({
    headers_to_add: z.array(HeaderSchema).optional(),
    headers_to_modify: z.array(HeaderSchema).optional(),
    headers_to_remove: z.array(HeaderRemoveSchema).optional(),
  }),
});

export type HeaderManipulationFormSchemaType = z.infer<
  typeof HeaderManipulationFormSchema
>;
