import { TemplateStringSeparator } from "./templateString";
import { z } from "zod";

/**
 * An ID is a string that is unique within a page and identifies a resource.
 * IDs are used when one resource needs to reference another resource, like
 * when one block references dynamic data from a query. The ID must not contain
 * any dots.
 */
export const Id = z
  .string()
  .min(1)
  .refine(
    (s) => !s.includes(TemplateStringSeparator),
    `IDs cannot contain a ${TemplateStringSeparator} as they are used as a separator for variables`
  );
