import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  oid: z.string(),
});

export const NamespaceListSchema = z.object({
  pageInfo: z.object({
    order: z.array(z.string()),
    filter: z.array(z.string()),
    limit: z.number(),
    offset: z.number(),
    total: z.number(),
  }),
  results: z.array(NamespaceSchema),
});
