import { z } from "zod";

const FiltersSchema = z.object({
  before: z.string().optional(),
  createdBefore: z.string().optional(),
  createdAfter: z.string().optional(),
  receivedBefore: z.string().optional(),
  receivedAfter: z.string().optional(),
  eventContains: z.string().optional(),
  typeContains: z.string().optional(),
});

export type FiltersSchemaType = z.infer<typeof FiltersSchema>;
