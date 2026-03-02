import { z } from "zod";

// when { { foo: "spam", somethingelse: false } };
export const RecordJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .object({
      Record: z.record(jsonExprSchema),
    })
    .strict();

export type RecordJsonExprSchemaType = z.infer<
  ReturnType<typeof RecordJsonExprSchema>
>;
