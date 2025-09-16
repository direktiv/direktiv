import { useMutation, useQueryClient } from "@tanstack/react-query";

import { LocalVariablesContent } from "../primitives/Variable/VariableContext";
import { MutationType } from "../../schema/procedures/mutation";
import { keyValueArrayToObject } from "../primitives/keyValue/utils";
import { useExtendedKeyValueArrayResolver } from "../primitives/keyValue/useExtenedKeyValueArrayResolver";
import { useKeyValueArrayResolver } from "../primitives/keyValue/useKeyValueArrayResolver";
import { useUrlGenerator } from "./utils";

type UsePageMutationParams = {
  onError?: (error: Error) => void;
};

export const usePageMutation = ({ onError }: UsePageMutationParams = {}) => {
  const queryClient = useQueryClient();
  const generateUrl = useUrlGenerator();
  const resolveKeyValueArray = useKeyValueArrayResolver();
  const resolveExtendedKeyValueArray = useExtendedKeyValueArrayResolver();

  return useMutation({
    mutationFn: async ({
      mutation,
      formVariables,
    }: {
      mutation: MutationType;
      formVariables: LocalVariablesContent;
    }) => {
      const { method, requestBody, requestHeaders } = mutation;

      const requestBodyResolved = resolveExtendedKeyValueArray(
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
