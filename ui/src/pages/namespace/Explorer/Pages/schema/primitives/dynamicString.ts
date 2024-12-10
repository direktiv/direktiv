import { z } from "zod";

/**
 * The schema and type of a variable string is just a normal string, but it
 * gets its own file to visualise that this string will get some special
 * processing when it is rendered on the page. It will allow you to include
 * special placeholders to use dynamic data from a parent context, such as
 * form data, data from an API query, etc.
 */
export const DynamicString = z.string().min(1);
