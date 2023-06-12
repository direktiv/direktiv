import { NodeListSchemaType } from "~/api/tree/schema";
import { RJSFSchema } from "@rjsf/utils";
import YAML from "js-yaml";
import { z } from "zod";

const validationSchema = z.object({
  type: z.literal("validate"),
  schema: z.object({ title: z.string() }).passthrough(),
});

const workflowSchema = z.object({
  states: z.array(validationSchema).nonempty(),
});

export const getValidationSchema = (node: NodeListSchemaType | undefined) => {
  const workflowData = node?.revision?.source && atob(node?.revision?.source);
  const workflowDataJson = YAML.load(workflowData ?? "");
  const parsed = workflowSchema.passthrough().safeParse(workflowDataJson);
  if (!parsed.success) return null;
  return parsed.data.states[0].schema as RJSFSchema;
};
