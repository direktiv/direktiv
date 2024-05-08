import { JqQueryResult, JqQueryResultType } from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { encode } from "js-base64";
import { getMessageFromApiError } from "~/api/errorHandling";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";

export type ExecuteJqueryPayload = {
  jq: string;
  data: string;
};

export const executeJquery = apiFactory<ExecuteJqueryPayload>({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/api/v2/jx`,
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
    mutationFn: ({ jq, data }: ExecuteJqueryPayload) =>
      executeJquery({
        apiKey: apiKey ?? undefined,
        urlParams: {},
        payload: {
          jq: encode(jq),
          data: encode(data),
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
