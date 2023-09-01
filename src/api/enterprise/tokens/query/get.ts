import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { TokenListSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { faker } from "@faker-js/faker";
import { set } from "date-fns";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

// TODO: remove the line below and delete the mock function
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const getTokens = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/tokens`,
  method: "GET",
  schema: TokenListSchema,
});

const getTokensMock = (_params: {
  apiKey?: string;
  urlParams: { namespace: string };
}): Promise<z.infer<typeof TokenListSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        tokens: [
          {
            id: faker.datatype.uuid(),
            description: faker.lorem.sentence(),
            permissions: [
              "opaManage",
              "variablesView",
              "registriesManage",
              "explorerManage",
              "registriesView",
              "nsconfigView",
              "eventsSend",
              "instancesView",
              "secretsView",
              "secretsManage",
              "servicesView",
              "servicesManage",
              "instancesManage",
              "explorerView",
              "workflowView",
              "workflowManage",
              "variablesManage",
              "nsconfigManage",
              "deleteNamespace",
              "eventsView",
              "workflowExecute",
              "workflowStore",
              "permissionsView",
              "permissionsManage",
              "opaView",
              "eventsManage",
            ],
            created: new Date().toISOString(),
            expires: set(new Date(), { year: 2022 }).toISOString(),
            expired: faker.datatype.boolean(),
          },
          {
            id: faker.datatype.uuid(),
            description: faker.lorem.sentence(),
            permissions: [],
            created: new Date().toISOString(),
            expires: set(new Date(), { year: 2022 }).toISOString(),
            expired: faker.datatype.boolean(),
          },
          {
            id: faker.datatype.uuid(),
            description: faker.lorem.sentence(),
            permissions: ["workflowView", "permissionsView"],
            created: new Date().toISOString(),
            expires: set(new Date(), { year: 2022 }).toISOString(),
            expired: faker.datatype.boolean(),
          },
        ],
      });
    }, 500);
  });

const fetchTokens = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof tokenKeys)["tokenList"]>>) =>
  getTokensMock({
    apiKey,
    urlParams: { namespace },
  });

export const useTokens = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: tokenKeys.tokenList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchTokens,
    enabled: !!namespace,
  });
};
