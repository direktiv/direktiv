import {
  QueryKey,
  UseQueryOptions,
  UseQueryResult,
  useQuery,
} from "@tanstack/react-query";

import { getPermissionStatus } from "./errorHandling";
import { useTranslation } from "react-i18next";

type UseQueryParam<
  TQueryFnData,
  TError,
  TData,
  TQueryKey extends QueryKey
> = UseQueryOptions<TQueryFnData, TError, TData, TQueryKey>;

/**
 * this type defines the additional properties that are added to the useQuery return value
 */
type ExtendedUseQueryReturn =
  | {
      isAllowed: false;
      noPermissionMessage: string;
    }
  | {
      isAllowed: true;
      noPermissionMessage: undefined;
    };

/**
 * useQueryWithPermissions is a wrapper around useQuery that will add permission handling to
 * the useQuery result. It checkes for the error and determines if the error is a permission
 * error.
 *
 * It behaves the same way as useQuery but has two additional properties:
 *
 * isAllowed: A boolean value that indicates whether the user has the necessary permissions
 * to proceed.
 *
 * noPermissionMessage: A string containing a message to display to the user when they lack
 * the required permissions. This message comes either from the API response or from a generic
 * message in the translation file if the API does not return any specific message. This propery
 * is undefined when isAllowed is true and always a string when isAllowed is false.
 */
const useQueryWithPermissions = <
  TQueryFnData = unknown,
  TError = unknown,
  TData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey
>(
  useQueryParams: UseQueryParam<TQueryFnData, TError, TData, TQueryKey>
): UseQueryResult<TData, TError> & ExtendedUseQueryReturn => {
  const { t } = useTranslation();
  const useQueryReturnValue = useQuery({
    ...useQueryParams,
  });

  const { error } = useQueryReturnValue;

  if (error) {
    const permissionStatus = getPermissionStatus(error);
    const isAllowed = permissionStatus.isAllowed;
    if (isAllowed === false) {
      const noPermissionMessage =
        permissionStatus.message ??
        t("api.generic.noPermissionQueryDescription");
      return { ...useQueryReturnValue, isAllowed, noPermissionMessage };
    }
  }

  return {
    ...useQueryReturnValue,
    isAllowed: true,
    noPermissionMessage: undefined,
  };
};

export default useQueryWithPermissions;
