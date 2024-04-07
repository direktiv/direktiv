import { RJSFSchema } from "@rjsf/utils";
import { parse } from "yaml";
import { z } from "zod";

export const workflowInputSchema = z.string().refine((string) => {
  try {
    JSON.parse(string);
    return true;
  } catch (error) {
    return false;
  }
});

const validationSchema = z.object({
  type: z.literal("validate"),
  schema: z.object({}).passthrough(), // allow any object and keep all the entries
});

const workflowSchema = z.object({
  // first step must be a validationSchema, the rest can be anything
  states: z.tuple([validationSchema]).rest(z.unknown()),
});

export const getValidationSchemaFromYaml = (
  workflowContent: string | undefined
) => {
  const workflowDataJson = parse(workflowContent ?? "");
  const parsed = workflowSchema.passthrough().safeParse(workflowDataJson);
  if (!parsed.success) return null;
  return parsed.data.states[0].schema as RJSFSchema;
};
