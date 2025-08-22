import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { LocalVariables } from "../primitives/Variable/LocalVariables";
import { MutationType } from "../../schema/procedures/mutation";
import { useUrlGenerator } from "./utils";

type UsePageMutationParams = {
  onError?: (error: Error) => void;
};

export const usePageMutation = ({ onError }: UsePageMutationParams = {}) => {
  const queryClient = useQueryClient();
  const generateUrl = useUrlGenerator();
  const resolveKeyValueArray = useKeyValueArrayResolver();

  return useMutation({
    mutationFn: async ({
      mutation,
      formVariables,
    }: {
      mutation: MutationType;
      formVariables: LocalVariables;
    }) => {
      const { method, requestBody, requestHeaders } = mutation;

      const requestBodyResolved = resolveKeyValueArray(
        requestBody ?? [],
        formVariables
      );

      const body = JSON.stringify(keyValueArrayToObject(requestBodyResolved));

      const requestHeadersResolved = resolveKeyValueArray(
        requestHeaders ?? [],
        formVariables
      );
      const headers = keyValueArrayToObject(requestHeadersResolved);

      const url = generateUrl(mutation, formVariables);

      const response = await fetch(url, {
        method,
        body,
        headers,
      });
      if (!response.ok) {
        throw new Error(`${response.status}: ${response.statusText}`);
      }
      return;
    },
    onSuccess: () => {
      queryClient.invalidateQueries();
    },
    onError: (error) => {
      onError?.(error);
    },
  });
};
