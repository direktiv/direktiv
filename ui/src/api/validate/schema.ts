import z from "zod";

const WorkflowValidationMessages = z.array(
  z.object({
    message: z.string().min(1),
    startLine: z.number(),
    startColumn: z.number(),
    endLine: z.number(),
    endColumn: z.number(),
    severity: z.enum(["hint", "info", "warning", "error"]),
  })
);

export type WorkflowValidationMessagesType = z.infer<
  typeof WorkflowValidationMessages
>;
