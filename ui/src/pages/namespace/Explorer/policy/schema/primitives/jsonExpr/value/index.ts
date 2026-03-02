import { z } from "zod";

// when { true };
export const ValueJsonExprSchema = z.object({ Value: z.unknown() }).strict();
