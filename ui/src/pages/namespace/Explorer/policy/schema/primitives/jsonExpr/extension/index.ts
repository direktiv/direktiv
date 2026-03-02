import { z } from "zod";

const ReservedJsonExprKeys = new Set([
  "Value",
  "Var",
  "Slot",
  "Unknown",
  "!",
  "neg",
  "isEmpty",
  "==",
  "!=",
  "in",
  "<",
  "<=",
  ">",
  ">=",
  "&&",
  "||",
  "+",
  "-",
  "*",
  "contains",
  "containsAll",
  "containsAny",
  "hasTag",
  "getTag",
  ".",
  "has",
  "is",
  "like",
  "if-then-else",
  "Set",
  "Record",
]);

// when { decimal("10.0") } / when { context.source_ip.isInRange(ip("222.222.222.0/24")) };
export const ExtensionJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .record(z.array(jsonExprSchema))
    .refine((value) => {
      const keys = Object.keys(value);
      return keys.length === 1;
    })
    .refine((value) => {
      const [key] = Object.keys(value);
      return key !== undefined && !ReservedJsonExprKeys.has(key);
    });

export type ExtensionJsonExprSchemaType = z.infer<
  ReturnType<typeof ExtensionJsonExprSchema>
>;
