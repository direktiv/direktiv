import {
  QueryKey,
  UseInfiniteQueryOptions,
  UseInfiniteQueryResult,
  useInfiniteQuery,
} from "@tanstack/react-query";

import { getPermissionStatus } from "./errorHandling";
import { useTranslation } from "react-i18next";

type UseInfiniteQueryParam<
  TQueryFnData,
  TError,
  TData,
  TQueryData,
  TQueryKey extends QueryKey
> = UseInfiniteQueryOptions<TQueryFnData, TError, TData, TQueryData, TQueryKey>;

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
  TError = unknown,
  TData = TQueryFnData,
  TQueryData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey
>(
  useInfiniteQueryParams: UseInfiniteQueryParam<
    TQueryFnData,
    TError,
    TData,
    TQueryData,
    TQueryKey
  >
): UseInfiniteQueryResult<TData, TError> & ExtendedUseInfiniteQueryReturn => {
  const { t } = useTranslation();

  // @ts-expect-error for some reason this is throwing a ts error. However, the return type
  // seems to be correct. This might be easier to fix with React Query 5 where the the overloads
  // are removed.
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
