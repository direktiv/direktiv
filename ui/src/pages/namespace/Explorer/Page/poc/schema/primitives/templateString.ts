import { z } from "zod";

/**
 * A template string is implemented as a normal string, but it has its schema
 * and type to visualize that this string will undergo special processing when
 * rendered on the page. It allows the user to include special placeholders to
 * utilize dynamic data from a parent context, such as form data, data from an
 * API query, etc.
 */
export const TemplateString = z.string().min(1);
