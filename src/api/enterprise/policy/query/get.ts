import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { PolicySchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { policyKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

// TODO: remove this
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const getPolicy = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespace/${namespace}/policy`,
  method: "GET",
  schema: PolicySchema,
});

const getPolicyMock = (_params: {
  apiKey?: string;
  urlParams: { namespace: string };
}): Promise<z.infer<typeof PolicySchema>> =>
  Promise.resolve(`package authorization

default allow = false

allow {
    input.method == "GET"
    input.path = ["customers", customerID]
    input.user_roles[_] = "admin"
}

allow {
    input.method == "GET"
    input.path = ["customers", customerID]
    input.user_roles[_] = "support"
}
`);

const fetchPolicy = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof policyKeys)["get"]>>) =>
  getPolicyMock({
    apiKey,
    urlParams: {
      namespace,
    },
  });

export const usePolicy = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: policyKeys.get(namespace, { apiKey: apiKey ?? undefined }),
    queryFn: fetchPolicy,
  });
};
