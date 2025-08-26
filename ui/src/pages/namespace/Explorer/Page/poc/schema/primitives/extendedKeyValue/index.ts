import { TemplateString } from "../templateString";
import z from "zod";

const StringSchema = z.object({
  type: z.literal("string"),
  value: TemplateString.min(1),
});

export const VariableSchema = z.object({
  type: z.literal("variable"),
  value: z.string(),
});

export const BooleanSchema = z.object({
  type: z.literal("boolean"),
  value: z.boolean(),
});

export const NumberSchema = z.object({
  type: z.literal("number"),
  value: z.number(),
});

const DataType = z.discriminatedUnion("type", [
  StringSchema,
  VariableSchema,
  BooleanSchema,
  NumberSchema,
]);

/**
 * An extended key-value pair that supports multiple data types for the value,
 * including strings, variables, booleans, arrays, and objects.
 */
export const ExtendedKeyValue = z.object({
  key: z.string().min(1),
  value: DataType,
});

export type ExtendedKeyValueType = z.infer<typeof ExtendedKeyValue>;

export type ValueType = ExtendedKeyValueType["value"]["type"];
