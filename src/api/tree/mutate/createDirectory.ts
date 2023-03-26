import { TreeFolderCreatedSchema, TreeListSchemaType } from "../schema";
import { apiFactory, defaultKeys } from "../../utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { forceSlashIfPath } from "../utils";
import { namespaceKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../componentsNext/Toast";

const createDirectory = apiFactory({
  pathFn: ({
    namespace,
    path,
    directory,
  }: {
    namespace: string;
    path?: string;
    directory: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceSlashIfPath(
      path
    )}/${directory}?op=create-directory`,
  method: "PUT",
  schema: TreeFolderCreatedSchema,
});

export const useCreateDirectory = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path, directory }: { path?: string; directory: string }) =>
      createDirectory({
        apiKey: apiKey ?? undefined,
        params: undefined,
        pathParams: {
          directory,
          namespace: namespace,
          path,
        },
      }),

    onSuccess(data, variables) {
      toast({
        title: "Directory created",
        description: `Directory ${variables.directory} was created in ${variables.path}`,
        variant: "success",
      });

      queryClient.setQueryData<TreeListSchemaType>(
        namespaceKeys.all(
          apiKey ?? defaultKeys.apiKey,
          namespace,
          variables.path ?? ""
        ),
        (oldData) => {
          const oldChildren = oldData?.children;
          return {
            ...oldData,
            children: {
              // may remove page info, since we will do all pagination locally anyways and this his hard to keep in sync
              pageInfo: oldChildren?.pageInfo
                ? {
                    ...oldChildren?.pageInfo,
                    total: oldChildren?.pageInfo.total + 1,
                  }
                : {
                    filter: [],
                    limit: 0,
                    offset: 0,
                    order: [],
                    total: 1,
                  },
              results: [...(oldChildren?.results ?? []), data.node],
            },
          };
        }
      );
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create directory 😢",
        variant: "error",
      });
    },
  });
};
