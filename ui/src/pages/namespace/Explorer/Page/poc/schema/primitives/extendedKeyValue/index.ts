import {
  BooleanArraySchema,
  NumberArraySchema,
  StringArraySchema,
} from "./array";

import { FlatObjectSchema } from "./object";
import z from "zod";

const StringSchema = z.object({
  type: z.literal("string"),
  value: z.string(),
});

const VariableSchema = z.object({
  type: z.literal("variable"),
  value: z.string(),
});

const BooleanSchema = z.object({
  type: z.literal("boolean"),
  value: z.boolean(),
});

const DataType = z.discriminatedUnion("type", [
  StringSchema,
  VariableSchema,
  BooleanSchema,
  StringArraySchema,
  BooleanArraySchema,
  NumberArraySchema,
  FlatObjectSchema,
]);

/**
 * An extended key-value pair that supports multiple data types for the value,
 * including strings, variables, booleans, arrays, and objects.
 */
export const ExtendedKeyValueSchema = z.object({
  key: z.string().min(1),
  value: DataType,
});
