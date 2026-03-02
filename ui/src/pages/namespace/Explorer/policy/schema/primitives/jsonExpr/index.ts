import { ValueJsonExprSchema } from "./value";
import { z } from "zod";

export const JsonExprSchema = ValueJsonExprSchema;

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
