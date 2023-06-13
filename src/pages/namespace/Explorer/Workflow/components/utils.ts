import { RJSFSchema } from "@rjsf/utils";
import YAML from "js-yaml";
import { z } from "zod";

const validationSchema = z.object({
  type: z.literal("validate"),
  schema: z.object({}).passthrough(), // allow any object and keep all the entries
});

const workflowSchema = z.object({
  states: z.array(validationSchema).nonempty(),
});

export const getValidationSchema = (workflowContent: string | undefined) => {
  const workflowDataJson = YAML.load(workflowContent ?? "");
  const parsed = workflowSchema.passthrough().safeParse(workflowDataJson);
  if (!parsed.success) return null;
  return parsed.data.states[0].schema as RJSFSchema;
};
