import type { QueryFunctionContext } from "@tanstack/react-query";
import { WorkflowVariableListSchema } from "../schema/workflowVariable";
import { apiFactory } from "~/api/apiFactory";
import { forceLeadingSlash } from "../utils";
import { treeKeys } from "../";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getVariables = apiFactory({
  url: ({ namespace, path }: { namespace: string; path?: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(path)}?op=vars`,
  method: "GET",
  schema: WorkflowVariableListSchema,
});

const fetchVariables = async ({
  queryKey: [{ apiKey, namespace, path }],
}: QueryFunctionContext<
  ReturnType<(typeof treeKeys)["workflowVariablesList"]>
>) =>
  getVariables({
    apiKey,
    urlParams: {
      namespace,
      path,
    },
  });

export const useWorkflowVariables = ({
  path,
  namespace: givenNamespace,
}: {
  path: string;
  namespace?: string | null;
}) => {
  const apiKey = useApiKey();
  const defaultNamespace = useNamespace();

  const namespace = givenNamespace ? givenNamespace : defaultNamespace;

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  if (!path) {
    throw new Error("path is undefined");
  }

  return useQuery({
    queryKey: treeKeys.workflowVariablesList(namespace, {
      apiKey: apiKey ?? undefined,
      path,
    }),
    queryFn: fetchVariables,
    enabled: !!namespace,
  });
};
