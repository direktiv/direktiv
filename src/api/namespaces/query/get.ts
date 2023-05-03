import { NamespaceListSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "../../utils";
import { namespaceKeys } from "../";
import { useApiKey } from "../../../util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

const getNamespaces = apiFactory({
  pathFn: () => `/api/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

const fetchNamespaces = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey: apiKey,
    params: undefined,
    pathParams: undefined,
  });

export const useListNamespaces = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: namespaceKeys.all(apiKey ?? undefined),
    queryFn: fetchNamespaces,
  });
};
