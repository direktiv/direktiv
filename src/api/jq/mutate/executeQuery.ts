import { JqQueryResult, JqQueryResultType } from "../schema";
import {
  getMessageFromApiError,
  useMutationWithPermissionHandling,
} from "~/api/errorHandling";

import { apiFactory } from "~/api/apiFactory";
import { useApiKey } from "~/util/store/apiKey";

export const executeJquery = apiFactory({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/api/jq`,
  method: "POST",
  schema: JqQueryResult,
});

export const useExecuteJQuery = ({
  onSuccess,
  onError,
}: {
  onSuccess?: (data: JqQueryResultType) => void;
  onError?: (error?: string) => void;
} = {}) => {
  const apiKey = useApiKey();
  return useMutationWithPermissionHandling({
    mutationFn: ({
      query,
      inputJsonString,
    }: {
      query: string;
      inputJsonString: string;
    }) =>
      executeJquery({
        apiKey: apiKey ?? undefined,
        urlParams: {},
        payload: {
          query,
          data: btoa(inputJsonString),
        },
      }),
    onSuccess: (res) => {
      onSuccess?.(res);
    },
    onError: (e) => {
      onError?.(getMessageFromApiError(e));
    },
  });
};
