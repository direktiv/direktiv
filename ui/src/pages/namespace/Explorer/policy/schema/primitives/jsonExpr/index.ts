import { AttributeJsonExprSchema } from "./attribute";
import { BinaryJsonExprSchema } from "./binary";
import { ExtensionJsonExprSchema } from "./extension";
import { IfThenElseJsonExprSchema } from "./ifThenElse";
import { IsJsonExprSchema } from "./is";
import { LikeJsonExprSchema } from "./like";
import { RecordJsonExprSchema } from "./record";
import { SetJsonExprSchema } from "./set";
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
    IfThenElseJsonExprSchema(JsonExprSchema),
    SetJsonExprSchema(JsonExprSchema),
    RecordJsonExprSchema(JsonExprSchema),
    ExtensionJsonExprSchema(JsonExprSchema),
  ])
);

export type JsonExprSchemaType = z.infer<typeof JsonExprSchema>;
