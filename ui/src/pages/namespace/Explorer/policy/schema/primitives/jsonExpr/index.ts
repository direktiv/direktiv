import { SlotJsonExprSchema } from "./slot";
import { VarJsonExprSchema } from "./var";
import { ValueJsonExprSchema } from "./value";
import { z } from "zod";

export const JsonExprSchema = z.union([
  ValueJsonExprSchema,
  VarJsonExprSchema,
  SlotJsonExprSchema,
]);

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
