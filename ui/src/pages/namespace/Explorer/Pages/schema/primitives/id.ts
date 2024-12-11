import { z } from "zod";

/**
 * An ID is a string that is unique within a page and identifies a resource.
 * IDs are used when one resource needs to reference another resource, like
 * when one block references dynamic data from a query.
 */
export const Id = z.string().min(1);
