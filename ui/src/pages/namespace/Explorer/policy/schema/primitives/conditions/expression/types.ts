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

// leaf nodes that do not contain child expressions.
type NonRecursiveExpression =
  | ValueExpression
  | VarExpression
  | SlotExpression
  | UnknownExpression;

// the input-side version of the same leaf-node union.
// It lets the recursive schema describe what it accepts
// separately from the parsed output it returns.
type NonRecursiveExpressionInput =
  | ValueExpressionInput
  | VarExpressionInput
  | SlotExpressionInput
  | UnknownExpressionInput;

// Reusable payload for Cedar's `is` expression.
// The generic keeps it reusable for both output types and input types.
type IsPayload<TExpression> = {
  left: TExpression;
  entity_type: string;
  in?: TExpression;
};

// Reusable payload for `like`
type LikePayload<TExpression> = {
  left: TExpression;
  pattern: PatternElement[];
};

// Reusable payload for `if-then-else`
type IfThenElsePayload<TExpression> = {
  if: TExpression;
  then: TExpression;
  else: TExpression;
};

// These types describe the recursive expression nodes on the output side.
// Each child points back to ExpressionType, which is what makes the tree
// recursive at the type level.

// unary expression like { "!": { arg: ... } }
type UnaryExpressionType = SingleKeyExpression<
  UnaryOperator,
  { arg: ExpressionType }
>;

// binary expression like { "==": { left: ..., right: ... } }.
type BinaryExpressionType = SingleKeyExpression<
  BinaryOperator,
  { left: ExpressionType; right: ExpressionType }
>;

// attribute expression like { ".": ... }
type AttributeExpressionType = SingleKeyExpression<
  AttributeOperator,
  { left: ExpressionType; attr: string }
>;

type IsExpressionType = { is: IsPayload<ExpressionType> };

type LikeExpressionType = { like: LikePayload<ExpressionType> };

type IfThenElseExpressionType = {
  "if-then-else": IfThenElsePayload<ExpressionType>;
};

// Set literal whose elements are themselves expressions.
type SetExpressionType = { Set: ExpressionType[] };

// Record literal whose property values are themselves expressions.
type RecordExpressionType = { Record: Record<string, ExpressionType> };

// This is the output-side union used by ExpressionSchemaType.
// It represents the recursive expression tree after parsing.
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

// The input-side recursive aliases mirror the output-side ones, but each child
// refers back to ExpressionInputType because nested input can also be recursive.

// Unary input node before parsing.
type UnaryExpressionInputType = SingleKeyExpression<
  UnaryOperator,
  { arg: ExpressionInputType }
>;

// Binary input node before parsing.
type BinaryExpressionInputType = SingleKeyExpression<
  BinaryOperator,
  { left: ExpressionInputType; right: ExpressionInputType }
>;

// Attribute input node before parsing.
type AttributeExpressionInputType = SingleKeyExpression<
  AttributeOperator,
  { left: ExpressionInputType; attr: string }
>;

// Input-side `is` node.
type IsExpressionInputType = { is: IsPayload<ExpressionInputType> };

// Input-side `like` node.
type LikeExpressionInputType = { like: LikePayload<ExpressionInputType> };

// Input-side conditional node.
type IfThenElseExpressionInputType = {
  "if-then-else": IfThenElsePayload<ExpressionInputType>;
};

// Input-side set literal.
type SetExpressionInputType = { Set: ExpressionInputType[] };

// Input-side record literal.
type RecordExpressionInputType = {
  Record: Record<string, ExpressionInputType>;
};

// Input-side extension call, keyed by a valid custom extension name and using
// recursive expression inputs as its argument list.
type ExtensionExpressionInputType = {
  [Key in ExtensionIdentifier]?: ExpressionInputType[];
};

// This is the full input union for the recursive schema.
// ExpressionSchemaType uses it to describe the values Zod is allowed to accept.
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

// This is the type-level contract for ExpressionSchema.
// It says the runtime schema must accept ExpressionInputType and parse it into
// ExpressionType, which is how we keep the recursive schema typed without using
// z.any() for self-references.
export type ExpressionSchemaType = z.ZodType<
  ExpressionType,
  z.ZodTypeDef,
  ExpressionInputType
>;
