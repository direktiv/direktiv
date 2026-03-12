import {
  type AttributeOperator,
  type BinaryOperator,
  type UnaryOperator,
} from "./utils";
import {
  ExtensionExpressionSchema,
  type ExtensionIdentifier,
} from "./extension";
import { LikeExpressionSchema, type PatternElement } from "./like";
import {
  type SlotExpression,
  type SlotExpressionInput,
  SlotExpressionSchema,
} from "./slot";
import {
  type UnknownExpression,
  type UnknownExpressionInput,
  UnknownExpressionSchema,
} from "./unknown";
import {
  type ValueExpression,
  type ValueExpressionInput,
  ValueExpressionSchema,
} from "./value";
import {
  type VarExpression,
  type VarExpressionInput,
  VarExpressionSchema,
} from "./var";

import { AttributeExpressionSchema } from "./attribute";
import { BinaryExpressionSchema } from "./binary";
import { IfThenElseExpressionSchema } from "./ifThenElse";
import { IsExpressionSchema } from "./is";
import { RecordExpressionSchema } from "./record";
import { SetExpressionSchema } from "./set";
import { UnaryExpressionSchema } from "./unary";
import { z } from "zod";

// Turns a union of operator names like "==" | ">" into a union of
// single-property objects such as { "==": ... } | { ">": ... }.
type SingleKeyExpression<Key extends string, Value> = {
  [K in Key]: { [P in K]: Value };
}[Key];

// Non recursive expressions
type NonRecursiveExpression =
  | ValueExpression
  | VarExpression
  | SlotExpression
  | UnknownExpression;

// describes what each schema accepts before parsing. We keep
// input and output types separate so the recursive schema can
// stay precise about both sides of the ZodType contract.
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

// This block defines the recursive output type for parsed expressions.
// We write the recursive TypeScript shape explicitly so we can annotate the
// schema with it, instead of falling back to z.any() for self-references.
// That breaks the schema/type circular dependency: the type exists first, then
// the lazy schema can declare that it parses into that type.
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

// This is the fully parsed expression tree type
type ExpressionType =
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
// Keeping the input type explicit lets the schema stay type-safe in both
// directions: what it accepts and what it returns after parsing.
type ExpressionInputType =
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

export type ExpressionSchemaType = z.ZodType<
  ExpressionType,
  z.ZodTypeDef,
  ExpressionInputType
>;

// z.lazy is what makes the recursive schema possible at runtime.
// Without it, ExpressionSchema would try to reference itself before the value
// exists. ExpressionSchemaType defines the contract for this recursive schema:
// it must accept ExpressionInputType and parse it into ExpressionType.
export const ExpressionSchema: ExpressionSchemaType = z.lazy(() =>
  z.union([
    ValueExpressionSchema,
    VarExpressionSchema,
    SlotExpressionSchema,
    UnknownExpressionSchema,
    UnaryExpressionSchema(ExpressionSchema),
    BinaryExpressionSchema(ExpressionSchema),
    AttributeExpressionSchema(ExpressionSchema),
    IsExpressionSchema(ExpressionSchema),
    LikeExpressionSchema(ExpressionSchema),
    IfThenElseExpressionSchema(ExpressionSchema),
    SetExpressionSchema(ExpressionSchema),
    RecordExpressionSchema(ExpressionSchema),
    ExtensionExpressionSchema(ExpressionSchema),
  ])
);
