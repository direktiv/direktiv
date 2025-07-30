import z from "zod";

const AllowedObjectValues = z.union([z.string(), z.number(), z.boolean()]);

// for simplifity we don't support nested objects yet
export const FlatObjectSchema = z.object({
  type: z.literal("object"),
  value: z.array(
    z.object({
      key: z.string().min(1),
      value: AllowedObjectValues,
    })
  ),
});
