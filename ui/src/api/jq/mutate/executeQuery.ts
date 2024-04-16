import { JqQueryResult, JqQueryResultType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { encode } from "js-base64";
import { getMessageFromApiError } from "~/api/errorHandling";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";

export const executeJquery = apiFactory<{ query: string; data: string }>({
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
  return useMutationWithPermissions({
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
          data: encode(inputJsonString),
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
