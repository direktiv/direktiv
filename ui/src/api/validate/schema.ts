import z from "zod";

export const WorkflowValidationSchema = z.array(
  z.object({
    message: z.string().min(1),
    startLine: z.number(),
    startColumn: z.number(),
    endLine: z.number(),
    endColumn: z.number(),
    severity: z.enum(["hint", "info", "warning", "error", ""]),
  })
);

export type WorkflowValidationSchemaType = z.infer<
  typeof WorkflowValidationSchema
>;
