import z from "zod";

export const StringArraySchema = z.object({
  type: z.literal("string-array"),
  value: z.array(z.string()),
});

export const BooleanArraySchema = z.object({
  type: z.literal("boolean-array"),
  value: z.array(z.boolean()),
});

export const NumberArraySchema = z.object({
  type: z.literal("number-array"),
  value: z.array(z.number()),
});
