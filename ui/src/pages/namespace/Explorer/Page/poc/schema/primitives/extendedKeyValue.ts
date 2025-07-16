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

// for simplifity we don't support nested arrays and objects yet
const AllowedArrayAndObjectValues = z.union([
  z.string(),
  z.number(),
  z.boolean(),
]);

const ArraySchema = z.object({
  type: z.literal("array"),
  value: z.array(AllowedArrayAndObjectValues),
});

const ObjectSchema = z.object({
  type: z.literal("object"),
  value: z.array(
    z.object({
      key: z.string().min(1),
      value: AllowedArrayAndObjectValues,
    })
  ),
});

const DataType = z.discriminatedUnion("type", [
  StringSchema,
  VariableSchema,
  BooleanSchema,
  ArraySchema,
  ObjectSchema,
]);

export const ExtendedKeyValueSchema = z.object({
  key: z.string().min(1),
  value: DataType,
});
