import { FormBase } from "./utils";
import { TemplateString } from "../../primitives/templateString";
import { z } from "zod";

export const FormTextarea = FormBase.extend({
  type: z.literal("form-textarea"),
  defaultValue: TemplateString.min(1),
});

export type FormTextareaType = z.infer<typeof FormTextarea>;
