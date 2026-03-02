import { AttributeJsonExprSchema } from "./attribute";
import { BinaryJsonExprSchema } from "./binary";
import { IsJsonExprSchema } from "./is";
import { LikeJsonExprSchema } from "./like";
import { SlotJsonExprSchema } from "./slot";
import { UnaryJsonExprSchema } from "./unary";
import { UnknownJsonExprSchema } from "./unknown";
import { ValueJsonExprSchema } from "./value";
import { VarJsonExprSchema } from "./var";
import { z } from "zod";

export const JsonExprSchema: z.ZodTypeAny = z.lazy(() =>
  z.union([
    ValueJsonExprSchema,
    VarJsonExprSchema,
    SlotJsonExprSchema,
    UnknownJsonExprSchema,
    UnaryJsonExprSchema(JsonExprSchema),
    BinaryJsonExprSchema(JsonExprSchema),
    AttributeJsonExprSchema(JsonExprSchema),
    IsJsonExprSchema(JsonExprSchema),
    LikeJsonExprSchema(JsonExprSchema),
  ])
);

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
