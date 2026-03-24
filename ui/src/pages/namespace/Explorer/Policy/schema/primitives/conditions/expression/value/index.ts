import { z } from "zod";

// when { context.request == {"riskScore": 92, "mfaPassed": true} };
export const ValueExpressionSchema = z.object({ Value: z.unknown() }).strict();

export type ValueExpression = z.infer<typeof ValueExpressionSchema>;
export type ValueExpressionInput = z.input<typeof ValueExpressionSchema>;
