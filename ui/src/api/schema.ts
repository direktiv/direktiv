import { z } from "zod";

export const PageinfoSchema = z.object({
  order: z.array(
    z.object({
      field: z.string(), // f.e. "NAME"
      direction: z.string(),
    })
  ),
  filter: z.array(
    z.object({
      field: z.string(), // f.e. "NAME"
      type: z.string(), // f.e. CONTAINS
      val: z.string(), // f.e. "something"
    })
  ),
  limit: z.number(),
  offset: z.number(),
  total: z.number(),
});

/**
 * FileSchema is an alternative to z.instanceof(File), since
 * Playwright throws a "File is not defined" error.
 */
export const FileSchema = z.custom<File>((value) => {
  if (
    typeof value === "object" &&
    value !== null &&
    "name" in value &&
    "size" in value &&
    "stream" in value
  ) {
    return true;
  } else {
    return false;
  }
});
