import { FormBase } from "./utils";
import { z } from "zod";

export const FormTextarea = FormBase.extend({
  type: z.literal("form-textarea"),
  defaultValue: z.string().optional(),
});

export type FormTextareaType = z.infer<typeof FormTextarea>;
