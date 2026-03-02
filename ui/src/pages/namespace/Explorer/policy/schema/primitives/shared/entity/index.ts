import { z } from "zod";

// User::"alice"
export const EntitySchema = z
  .object({
    type: z.string(),
    id: z.string(),
  })
  .strict();

type EntitySchemaType = z.infer<typeof EntitySchema>;
