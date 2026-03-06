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

export const ExpressionSchema: z.ZodTypeAny = z.lazy(() =>
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
