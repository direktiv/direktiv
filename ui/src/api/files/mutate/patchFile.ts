import {
  FilePatchedSchema,
  RenameFileSchemaType,
  UpdateFileSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "~/api/files/utils";

export const patchFile = apiFactory<
  RenameFileSchemaType | UpdateFileSchemaType
>({
  url: ({
    baseUrl,
    namespace,
    path,
  }: {
    baseUrl?: string;
    namespace: string;
    path: string;
  }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/files${forceLeadingSlash(
      path
    )}`,
  method: "PATCH",
  schema: FilePatchedSchema,
});
