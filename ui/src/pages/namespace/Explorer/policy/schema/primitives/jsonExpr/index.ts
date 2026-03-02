import { SlotJsonExprSchema } from "./slot";
import { UnaryJsonExprSchema } from "./unary";
import { UnknownJsonExprSchema } from "./unknown";
import { VarJsonExprSchema } from "./var";
import { ValueJsonExprSchema } from "./value";
import { z } from "zod";

export const JsonExprSchema: z.ZodTypeAny = z.lazy(() =>
  z.union([
    ValueJsonExprSchema,
    VarJsonExprSchema,
    SlotJsonExprSchema,
    UnknownJsonExprSchema,
    UnaryJsonExprSchema(JsonExprSchema),
  ])
);

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
