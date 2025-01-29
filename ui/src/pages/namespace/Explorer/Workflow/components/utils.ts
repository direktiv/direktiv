import { FileListSchemaType } from "~/api/files/schema";
import { RJSFSchema } from "@rjsf/utils";
import { decode } from "js-base64";
import { parse } from "yaml";
import { useTranslation } from "react-i18next";
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

export const useValidationSchemaFromFile = (
  file: FileListSchemaType["data"] | undefined
) => {
  const { t } = useTranslation();
  let result: RJSFSchema | null = null;
  let error: string | undefined;
  try {
    result =
      file?.type === "workflow"
        ? getValidationSchemaFromYaml(decode(file?.data ?? ""))
        : null;
  } catch (e) {
    error = t("pages.explorer.tree.workflow.runWorkflow.invalidYaml");
  }
  return { result, error };
};
