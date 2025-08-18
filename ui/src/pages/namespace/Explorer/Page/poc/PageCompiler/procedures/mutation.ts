import {
  keyValueArrayToObject,
  useKeyValueArrayResolver,
} from "../primitives/keyValue/utils";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { InjectedVariables } from "../primitives/Variable/VariableContext";
import { MutationType } from "../../schema/procedures/mutation";
import { useUrlGenerator } from "./utils";

export const usePageMutation = () => {
  const queryClient = useQueryClient();
  const generateUrl = useUrlGenerator();
  const resolveKeyValueArray = useKeyValueArrayResolver();

  return useMutation({
    mutationFn: async ({
      mutation,
      options,
    }: {
      mutation: MutationType;
      options?: { variables: InjectedVariables };
    }) => {
      const { method, requestBody, requestHeaders } = mutation;

      const requestBodyResolved = resolveKeyValueArray(
        requestBody ?? [],
        options
      );

      const body = JSON.stringify(keyValueArrayToObject(requestBodyResolved));

      const requestHeadersResolved = resolveKeyValueArray(
        requestHeaders ?? [],
        options
      );
      const headers = keyValueArrayToObject(requestHeadersResolved);

      const url = generateUrl(mutation, options);

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
  });
};
