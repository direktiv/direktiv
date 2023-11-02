import type { QueryFunctionContext } from "@tanstack/react-query";
import { RegistryListSchema } from "../schema";
import { apiFactory } from "../../apiFactory";
import { registriesKeys } from "..";
import { useApiKey } from "../../../util/store/apiKey";
import { useNamespace } from "../../../util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getRegistries = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/v2/namespaces/${namespace}/registries`,
  method: "GET",
  schema: RegistryListSchema,
});

const fetchRegistries = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<
  ReturnType<(typeof registriesKeys)["registriesList"]>
>) =>
  getRegistries({
    apiKey,
    urlParams: { namespace },
  });

export const useRegistries = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: registriesKeys.registriesList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchRegistries,
    enabled: !!namespace,
  });
};
