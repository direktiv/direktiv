import {
  DefaultError,
  UseMutationOptions,
  useMutation,
} from "@tanstack/react-query";

import { getPermissionStatus } from "./errorHandling";
import { t } from "i18next";
import { useToast } from "~/design/Toast";

type UseMutationParam<TData, TError, TVariables> = UseMutationOptions<
  TData,
  TError,
  TVariables
>;

/**
 * useMutationWithPermissions is a wrapper around useMutation that will hook
 * into the onError callback and check if the error is a permission error.
 * If it is, it will display a toast message to the user and early return.
 * So that no further error handling will be done.
 */
const useMutationWithPermissions = <
  TData = unknown,
  TError = DefaultError,
  TVariables = void
>(
  useMutationParams: UseMutationParam<TData, TError, TVariables>
) => {
  const { toast } = useToast();
  return useMutation({
    ...useMutationParams,
    onError: (error, variable, context) => {
      const res = getPermissionStatus(error);
      if (!res.isAllowed) {
        toast({
          title: t("api.generic.noPermissionTitle"),
          description:
            res.message ?? t("api.generic.noPermissionMutationDescription"),
          variant: "error",
        });
        return;
      }
      useMutationParams.onError?.(error, variable, context);
    },
  });
};

export default useMutationWithPermissions;
