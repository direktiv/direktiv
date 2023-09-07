import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { TokenListSchema } from "../schema";
import { faker } from "@faker-js/faker";
import moment from "moment";
import { tokenKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { z } from "zod";

// const getTokens = apiFactory({
//   url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
//     `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/tokens`,
//   method: "GET",
//   schema: TokenListSchema,
// });

// TODO: remove this mock
const getTokens = (_params: {
  apiKey?: string;
  urlParams: { namespace: string };
}): Promise<z.infer<typeof TokenListSchema>> =>
  new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        tokens: [
          {
            id: faker.datatype.uuid(),
            description: "some super token",
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
            expires: moment().add(1, "day").toISOString(),
            expired: false,
          },
          {
            id: faker.datatype.uuid(),
            description: "some useless token",
            permissions: [],
            created: new Date().toISOString(),
            expires: moment().add(1, "month").toISOString(),
            expired: false,
          },
          {
            id: faker.datatype.uuid(),
            description: "some expired token",
            permissions: ["workflowView", "permissionsView"],
            created: new Date().toISOString(),
            expires: moment().subtract(1, "day").toISOString(),
            expired: true,
          },
        ],
      });
    }, 500);
  });

const fetchTokens = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof tokenKeys)["tokenList"]>>) =>
  getTokens({
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
