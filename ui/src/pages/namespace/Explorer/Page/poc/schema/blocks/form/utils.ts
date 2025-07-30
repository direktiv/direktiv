import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

export const FormBase = z.object({
  id: z.string().min(1),
  label: TemplateString,
  description: TemplateString.optional(),
  required: z.boolean(),
});

export type FormBaseType = z.infer<typeof FormBase>;
