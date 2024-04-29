import {
  DefaultError,
  InfiniteData,
  QueryKey,
  UseInfiniteQueryOptions,
  UseInfiniteQueryResult,
  useInfiniteQuery,
} from "@tanstack/react-query";

import { getPermissionStatus } from "./errorHandling";
import { useTranslation } from "react-i18next";

/**
 * this type defines the additional properties that are added to the useInfiniteQuery return value
 */
type ExtendedUseInfiniteQueryReturn =
  | {
      isAllowed: false;
      noPermissionMessage: string;
    }
  | {
      isAllowed: true;
      noPermissionMessage: undefined;
    };

/**
 * see useQueryWithPermissions for more information. This is the exact same concept but for infinite queries.
 */
const useInfiniteQueryWithPermissions = <
  TQueryFnData = unknown,
  TError = DefaultError,
  TData = InfiniteData<TQueryFnData>,
  TQueryKey extends QueryKey = QueryKey,
  TPageParam = unknown
>(
  useInfiniteQueryParams: UseInfiniteQueryOptions<
    TQueryFnData,
    TError,
    TData,
    TQueryFnData,
    TQueryKey,
    TPageParam
  >
): UseInfiniteQueryResult<TData, TError> & ExtendedUseInfiniteQueryReturn => {
  const { t } = useTranslation();

  const useInfiniteQueryReturnValue = useInfiniteQuery({
    ...useInfiniteQueryParams,
  });
  const { error } = useInfiniteQueryReturnValue;

  if (error) {
    const permissionStatus = getPermissionStatus(error);
    const isAllowed = permissionStatus.isAllowed;
    if (isAllowed === false) {
      const noPermissionMessage =
        permissionStatus.message ??
        t("api.generic.noPermissionQueryDescription");
      return { ...useInfiniteQueryReturnValue, isAllowed, noPermissionMessage };
    }
  }

  return {
    ...useInfiniteQueryReturnValue,
    isAllowed: true,
    noPermissionMessage: undefined,
  };
};

export default useInfiniteQueryWithPermissions;
