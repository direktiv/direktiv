import {
  RegistryCreatedSchema,
  RegistryCreatedSchemaType,
  RegistryFormSchemaType,
} from "../schema";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/utils";
import { registriesKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useToast } from "~/design/Toast";

const createRegistry = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/functions/registries/namespaces/${namespace}`,
  method: "POST",
  schema: RegistryCreatedSchema,
});

export const useCreateRegistry = ({
  onSuccess,
}: {
  onSuccess?: (registry: RegistryCreatedSchemaType) => void;
} = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  const mutationFn = ({ url, user, password }: RegistryFormSchemaType) =>
    createRegistry({
      apiKey: apiKey ?? undefined,
      payload: { data: `${user}:${password}`, reg: url },
      urlParams: {
        namespace: namespace,
      },
      headers: undefined,
    });

  return useMutation({
    mutationFn,
    onSuccess: (registry, variables) => {
      queryClient.invalidateQueries(
        registriesKeys.registriesList(namespace, {
          apiKey: apiKey ?? undefined,
        })
      );
      toast({
        title: "Secret created",
        description: `Secret ${variables.url} was created.`,
        variant: "success",
      });
      onSuccess?.(null);
    },
  });
};
