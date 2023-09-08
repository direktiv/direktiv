import {
  QueryKey,
  UseMutationOptions,
  UseQueryOptions,
  UseQueryResult,
  useMutation,
  useQuery,
} from "@tanstack/react-query";

import { t } from "i18next";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";
import { z } from "zod";

/**
 * The ApiErrorSchema is a special schema we use to standardize api error handling
 * across the app. It contains the response object from the fetch api, and an
 * optional json object that may contain the error code and message. Since errors
 * are always typed as unknown, we can use the the custom type guard isApiErrorSchema
 * to check if an error conforms to the ApiErrorSchema and have typesafe way to process
 * the error.
 */
export const ApiErrorSchema = z.object({
  response: z.instanceof(Response),
  json: z
    .object({
      code: z.string().or(z.number()).optional(),
      message: z.string().optional(),
    })
    .passthrough()
    .optional(),
});

type ApiErrorSchemaType = z.infer<typeof ApiErrorSchema>;

export const createApiErrorFromResponse = async (
  res: Response
): Promise<ApiErrorSchemaType> => {
  let json: ApiErrorSchemaType["json"];
  try {
    json = await res.json();
  } catch (error) {
    process.env.NODE_ENV !== "test" && console.error(error);
  }

  return {
    response: res,
    json,
  };
};

export const isApiErrorSchema = (error: unknown): error is ApiErrorSchemaType =>
  ApiErrorSchema.safeParse(error).success;

export const getMessageFromApiError = (error: unknown) =>
  isApiErrorSchema(error) ? error.json?.message : undefined;

type PermissionStatus =
  | {
      isAllowed: true;
    }
  | {
      isAllowed: false;
      message?: string;
    };

export const getPermissionStatus = (error: unknown): PermissionStatus => {
  if (isApiErrorSchema(error)) {
    if (error.response.status === 401 || error.response.status === 403) {
      return {
        isAllowed: false,
        message: getMessageFromApiError(error),
      };
    }
  }

  return {
    isAllowed: true,
  };
};

type UseMutationParam<TData, TError, TVariables> = UseMutationOptions<
  TData,
  TError,
  TVariables
>;

/**
 * useMutationWithPermissionHandling is a wrapper around useMutation that will
 * hook into the onError callback and check if the error is a permission error.
 * If it is, it will display a toast message to the user and early return. So
 * that no further error handling will be done.
 */
export const useMutationWithPermissionHandling = <
  TData = unknown,
  TError = unknown,
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
          description: res.message ?? t("api.generic.noPermissionDescription"),
          variant: "error",
        });
        return;
      }
      useMutationParams.onError?.(error, variable, context);
    },
  });
};

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
export const useQueryWithPermissions = <
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
        permissionStatus.message ?? t("api.generic.noPermissionDescription");
      return { ...useQueryReturnValue, isAllowed, noPermissionMessage };
    }
  }

  return {
    ...useQueryReturnValue,
    isAllowed: true,
    noPermissionMessage: undefined,
  };
};
