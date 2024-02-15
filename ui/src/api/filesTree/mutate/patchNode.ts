import { NodePatchedSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/tree/utils";

export const patchNode = apiFactory({
  url: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/v2/namespaces/${namespace}/files-tree${forceLeadingSlash(path)}`,
  method: "PATCH",
  schema: NodePatchedSchema,
});
