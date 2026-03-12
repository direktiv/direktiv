import {
  type ExpressionInputType,
  type ExpressionSchemaType,
  type ExpressionType,
} from "./types";
import { AttributeExpressionSchema } from "./attribute";
import { BinaryExpressionSchema } from "./binary";
import { ExtensionExpressionSchema } from "./extension";
import { IfThenElseExpressionSchema } from "./ifThenElse";
import { IsExpressionSchema } from "./is";
import { LikeExpressionSchema } from "./like";
import { RecordExpressionSchema } from "./record";
import { SetExpressionSchema } from "./set";
import { SlotExpressionSchema } from "./slot";
import { UnaryExpressionSchema } from "./unary";
import { UnknownExpressionSchema } from "./unknown";
import { ValueExpressionSchema } from "./value";
import { VarExpressionSchema } from "./var";
import { z } from "zod";

export type { ExpressionInputType, ExpressionSchemaType, ExpressionType };

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
