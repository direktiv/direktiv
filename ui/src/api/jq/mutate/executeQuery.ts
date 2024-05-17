import {
  ExecuteJxQueryPayloadType,
  JxQueryResult,
  JxQueryResultType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { encode } from "js-base64";
import { getMessageFromApiError } from "~/api/errorHandling";
import { useApiKey } from "~/util/store/apiKey";
import useMutationWithPermissions from "~/api/useMutationWithPermissions";

export const executeJxQuery = apiFactory<ExecuteJxQueryPayloadType>({
  url: ({ baseUrl }: { baseUrl?: string }) => `${baseUrl ?? ""}/api/v2/jx`,
  method: "POST",
  schema: JxQueryResult,
});

export const useExecuteJxQuery = ({
  onSuccess,
  onError,
}: {
  onSuccess?: (data: JxQueryResultType) => void;
  onError?: (error?: string) => void;
} = {}) => {
  const apiKey = useApiKey();
  return useMutationWithPermissions({
    mutationFn: ({ data, jx }: ExecuteJxQueryPayloadType) =>
      executeJxQuery({
        apiKey: apiKey ?? undefined,
        urlParams: {},
        payload: {
          jx: encode(jx),
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
