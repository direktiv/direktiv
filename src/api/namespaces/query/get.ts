import { NamespaceListSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { namespaceKeys } from "../";
import { useApiKey } from "~/util/store/apiKey";
import { useQuery } from "@tanstack/react-query";

export const getNamespaces = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/api/namespaces`,
  method: "GET",
  schema: NamespaceListSchema,
});

const fetchNamespaces = async ({
  queryKey: [{ apiKey }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["all"]>>) =>
  getNamespaces({
    apiKey,
    urlParams: {},
  });

export const useListNamespaces = () => {
  const apiKey = useApiKey();
  return useQuery({
    queryKey: namespaceKeys.all(apiKey ?? undefined),
    queryFn: fetchNamespaces,
  });
};
