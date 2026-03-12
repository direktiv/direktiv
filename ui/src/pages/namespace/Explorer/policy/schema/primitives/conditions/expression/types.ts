import type { AttributeOperator, BinaryOperator, UnaryOperator } from "./utils";
import type { SlotExpression, SlotExpressionInput } from "./slot";
import type { UnknownExpression, UnknownExpressionInput } from "./unknown";
import type { ValueExpression, ValueExpressionInput } from "./value";
import type { VarExpression, VarExpressionInput } from "./var";
import type { ExtensionIdentifier } from "./extension";
import type { PatternElement } from "./like";
import { z } from "zod";

// Turns a union of operator names like "==" | ">" into a union of
// single-property objects such as { "==": ... } | { ">": ... }.
type SingleKeyExpression<Key extends string, Value> = {
  [K in Key]: { [P in K]: Value };
}[Key];

// Non-recursive expressions can reuse types exported by their leaf schemas.
// The recursive types below compose those leaves into the full expression tree.
type NonRecursiveExpression =
  | ValueExpression
  | VarExpression
  | SlotExpression
  | UnknownExpression;

// Input and output stay separate so ExpressionSchemaType can describe both what
// the recursive schema accepts and what it returns after parsing.
type NonRecursiveExpressionInput =
  | ValueExpressionInput
  | VarExpressionInput
  | SlotExpressionInput
  | UnknownExpressionInput;

type IsPayload<TExpression> = {
  left: TExpression;
  entity_type: string;
  in?: TExpression;
};

type LikePayload<TExpression> = {
  left: TExpression;
  pattern: PatternElement[];
};

type IfThenElsePayload<TExpression> = {
  if: TExpression;
  then: TExpression;
  else: TExpression;
};

// We write the recursive TypeScript shape explicitly so the runtime schema can
// refer to ExpressionSchemaType instead of falling back to z.any().
type UnaryExpressionType = SingleKeyExpression<
  UnaryOperator,
  { arg: ExpressionType }
>;

type BinaryExpressionType = SingleKeyExpression<
  BinaryOperator,
  { left: ExpressionType; right: ExpressionType }
>;

type AttributeExpressionType = SingleKeyExpression<
  AttributeOperator,
  { left: ExpressionType; attr: string }
>;

type IsExpressionType = { is: IsPayload<ExpressionType> };

type LikeExpressionType = { like: LikePayload<ExpressionType> };

type IfThenElseExpressionType = {
  "if-then-else": IfThenElsePayload<ExpressionType>;
};

type SetExpressionType = { Set: ExpressionType[] };

type RecordExpressionType = { Record: Record<string, ExpressionType> };

// This is the fully parsed expression tree.
export type ExpressionType =
  | NonRecursiveExpression
  | UnaryExpressionType
  | BinaryExpressionType
  | AttributeExpressionType
  | IsExpressionType
  | LikeExpressionType
  | IfThenElseExpressionType
  | SetExpressionType
  | RecordExpressionType;

type UnaryExpressionInputType = SingleKeyExpression<
  UnaryOperator,
  { arg: ExpressionInputType }
>;

type BinaryExpressionInputType = SingleKeyExpression<
  BinaryOperator,
  { left: ExpressionInputType; right: ExpressionInputType }
>;

type AttributeExpressionInputType = SingleKeyExpression<
  AttributeOperator,
  { left: ExpressionInputType; attr: string }
>;

type IsExpressionInputType = { is: IsPayload<ExpressionInputType> };

type LikeExpressionInputType = { like: LikePayload<ExpressionInputType> };

type IfThenElseExpressionInputType = {
  "if-then-else": IfThenElsePayload<ExpressionInputType>;
};

type SetExpressionInputType = { Set: ExpressionInputType[] };

type RecordExpressionInputType = {
  Record: Record<string, ExpressionInputType>;
};

type ExtensionExpressionInputType = {
  [Key in ExtensionIdentifier]?: ExpressionInputType[];
};

// This mirrors the same recursive structure for schema inputs.
export type ExpressionInputType =
  | NonRecursiveExpressionInput
  | UnaryExpressionInputType
  | BinaryExpressionInputType
  | AttributeExpressionInputType
  | IsExpressionInputType
  | LikeExpressionInputType
  | IfThenElseExpressionInputType
  | SetExpressionInputType
  | RecordExpressionInputType
  | ExtensionExpressionInputType;

// This is the type-level contract for the recursive runtime schema.
export type ExpressionSchemaType = z.ZodType<
  ExpressionType,
  z.ZodTypeDef,
  ExpressionInputType
>;
