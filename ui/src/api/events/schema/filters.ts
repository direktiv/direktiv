import { z } from "zod";

const FiltersSchema = z.object({
  before: z.date().optional(),
  createdBefore: z.date().optional(),
  createdAfter: z.date().optional(),
  receivedBefore: z.date().optional(),
  receivedAfter: z.date().optional(),
  eventContains: z.string().optional(),
  typeContains: z.string().optional(),
});

export type FiltersSchemaType = z.infer<typeof FiltersSchema>;
