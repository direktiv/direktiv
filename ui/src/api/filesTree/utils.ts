import { FileSchemaType } from "./schema";

export const sortFoldersFirst = (
  a: FileSchemaType,
  b: FileSchemaType
): number => {
  if (a.type === "directory" && b.type !== "directory") {
    return -1;
  }

  if (b.type === "directory" && a.type !== "directory") {
    return 1;
  }

  return a.path.localeCompare(b.path);
};
