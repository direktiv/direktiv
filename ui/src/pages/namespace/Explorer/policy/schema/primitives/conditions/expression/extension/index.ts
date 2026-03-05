import { ExpressionReservedKeys } from "../utils";
import { z } from "zod";

/*
  when { decimal("100.00") <= context.invoiceAmount }
  when { context.source_ip.isInRange(ip("10.0.0.0/8")) };
*/
export const ExtensionExpressionSchema = (expressionSchema: z.ZodTypeAny) =>
  z
    .record(z.array(expressionSchema))
    .refine((value) => {
      const keys = Object.keys(value);
      return keys.length === 1;
    })
    .refine((value) => {
      const [key] = Object.keys(value);
      return key !== undefined && !ExpressionReservedKeys.has(key);
    });
