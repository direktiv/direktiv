import { PageinfoSchema } from "../schema";
import { z } from "zod";

export const NamespaceSchema = z.object({
  createdAt: z.string(),
  updatedAt: z.string(),
  name: z.string(),
  oid: z.string(),
});

export const NamespaceListSchema = z.object({
  pageInfo: PageinfoSchema,
  results: z.array(NamespaceSchema),
});

export const NamespaceCreatedSchema = z.object({
  namespace: NamespaceSchema,
});
