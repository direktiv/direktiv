import { NamespaceListSchema } from "./schema";
import { apiFactory } from "../utils";

export const getNamespaces = apiFactory({
  path: `/api/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

// export const versionKeys = {
//   all: ["versions"] as const,
// };
