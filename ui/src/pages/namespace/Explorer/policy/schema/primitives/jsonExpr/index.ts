import { SlotJsonExprSchema } from "./slot";
import { UnknownJsonExprSchema } from "./unknown";
import { VarJsonExprSchema } from "./var";
import { ValueJsonExprSchema } from "./value";
import { z } from "zod";

export const JsonExprSchema = z.union([
  ValueJsonExprSchema,
  VarJsonExprSchema,
  SlotJsonExprSchema,
  UnknownJsonExprSchema,
]);

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
