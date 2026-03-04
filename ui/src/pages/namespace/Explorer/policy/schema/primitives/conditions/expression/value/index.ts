import { z } from "zod";

// when { true };
export const ValueExpressionSchema = z.object({ Value: z.unknown() }).strict();
