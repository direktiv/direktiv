import type { QueryFunctionContext } from "@tanstack/react-query";
import { RegistryListSchema } from "../schema";
import { apiFactory } from "../../utils";
import { registriesKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import { useQuery } from "@tanstack/react-query";

const getRegistries = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/functions/registries/namespaces/${namespace}`,
  method: "GET",
  schema: RegistryListSchema,
});

const fetchRegistries = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<
  ReturnType<(typeof registriesKeys)["registriesList"]>
>) =>
  getRegistries({
    apiKey: apiKey,
    urlParams: { namespace },
    payload: undefined,
    headers: undefined,
  });

export const useRegistries = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: registriesKeys.registriesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchRegistries,
    enabled: !!namespace,
  });
};
