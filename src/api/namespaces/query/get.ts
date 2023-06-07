import { NamespaceListSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/utils";
import { namespaceKeys } from "../";
import { useApiKey } from "~/util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

const getNamespaces = apiFactory({
  url: () => `/api/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

const fetchNamespaces = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey,
    payload: undefined,
    headers: undefined,
    urlParams: undefined,
  });

export const useListNamespaces = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: namespaceKeys.all(apiKey ?? undefined),
    queryFn: fetchNamespaces,
  });
};
