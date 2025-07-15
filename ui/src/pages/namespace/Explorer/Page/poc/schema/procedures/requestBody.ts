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

const ArraySchema = z.object({
  type: z.literal("array"),
  value: z.array(z.unknown()),
});

const ObjectSchema = z.object({
  type: z.literal("object"),
  value: z.array(
    z.object({
      key: z.string().min(1),
      value: z.unknown(),
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

export const RequestBodySchema = z.object({
  key: z.string().min(1),
  value: DataType,
});
