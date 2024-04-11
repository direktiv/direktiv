import type { QueryFunctionContext } from "@tanstack/react-query";
import { VarListSchema } from "../schema";
import { apiFactory } from "../../apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { forceLeadingSlash } from "~/api/files/utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { varKeys } from "..";

type GetVarsParams = {
  namespace: string;
  workflowPath?: string;
};

const getVars = apiFactory({
  url: ({ namespace, ...queryParams }: GetVarsParams) => {
    const queryParamsString = buildSearchParamsString({
      ...queryParams,
    });

    return `/api/v2/namespaces/${namespace}/variables${queryParamsString}`;
  },
  method: "GET",
  schema: VarListSchema,
});

const fetchVars = async ({
  queryKey: [{ apiKey, namespace, workflowPath }],
}: QueryFunctionContext<ReturnType<(typeof varKeys)["varList"]>>) =>
  getVars({
    apiKey,
    urlParams: { namespace, workflowPath },
  });

type UseVarsParams = {
  namespace?: string;
  workflowPath?: string;
};

export const useVars = ({
  namespace: givenNamespace,
  workflowPath,
}: UseVarsParams = {}) => {
  const apiKey = useApiKey();
  const defaultNamespace = useNamespace();

  const namespace = givenNamespace ?? defaultNamespace;

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: varKeys.varList(namespace, {
      apiKey: apiKey ?? undefined,
      workflowPath: workflowPath ? forceLeadingSlash(workflowPath) : undefined,
    }),
    queryFn: fetchVars,
    enabled: !!namespace,
  });
};
