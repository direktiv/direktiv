import { z } from "zod";

/**
 * A dynamic string is implemented just a normal string, but it gets its schema
 * and type to visualise that this string will get some special processing when
 * they are rendered on the page. It will allow the user to include special
 * placeholders to use dynamic data from a parent context, such as form data,
 * data from an API query, etc.
 */
export const DynamicString = z.string().min(1);
