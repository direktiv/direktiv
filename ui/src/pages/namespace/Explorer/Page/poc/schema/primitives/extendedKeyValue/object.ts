import z from "zod";

// for simplifity we don't support nested objects yet
const AllowedObjectValues = z.union([z.string(), z.number(), z.boolean()]);

export const ObjectSchema = z.object({
  type: z.literal("object"),
  value: z.array(
    z.object({
      key: z.string().min(1),
      value: AllowedObjectValues,
    })
  ),
});
