import {
  type BinaryOperator,
  ExpressionReservedKeys,
  type LowercaseLetter,
  type UnaryOperator,
} from "../utils";
import type { ExpressionSchemaType } from "../types";
import { z } from "zod";

type ReservedExtensionIdentifier =
  | UnaryOperator
  | BinaryOperator
  | "has"
  | "is"
  | "like"
  | "if-then-else";

// Extension functions must start with a lowercase letter and must not collide
// with built-in Cedar expression keys/operators.
export type ExtensionIdentifier = Exclude<
  `${LowercaseLetter}${string}`,
  ReservedExtensionIdentifier
>;

/*
  when { decimal("100.00") <= context.invoiceAmount }
  when { context.source_ip.isInRange(ip("10.0.0.0/8")) };
*/
export const ExtensionExpressionSchema = (
  expressionSchema: ExpressionSchemaType
) =>
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
