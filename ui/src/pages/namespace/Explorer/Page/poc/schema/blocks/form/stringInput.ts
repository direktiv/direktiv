import { FormBase } from "./utils";
import { z } from "zod";

export const stringInputTypes = ["text", "password", "email", "url"] as const;

export const FormStringInput = FormBase.extend({
  type: z.literal("form-string-input"),
  variant: z.enum(stringInputTypes),
  defaultValue: z.string(),
});

export type FormStringInputType = z.infer<typeof FormStringInput>;
