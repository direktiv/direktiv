import { VarJsonExprSchema } from "./var";
import { ValueJsonExprSchema } from "./value";
import { z } from "zod";

export const JsonExprSchema = z.union([ValueJsonExprSchema, VarJsonExprSchema]);

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
