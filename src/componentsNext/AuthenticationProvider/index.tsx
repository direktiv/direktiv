import { FC, PropsWithChildren, useEffect } from "react";
import { useApiActions, useApiKey } from "~/util/store/apiKey";

import { Authdialog } from "../Authdialog";
import useApiKeyHandling from "~/hooksNext/useApiKeyHandling";
import { useRefreshSession } from "~/api/enterprise/session/query/ping";

export const AuthenticationProvider: FC<PropsWithChildren> = ({ children }) => {
  const { isFetched, isCurrentKeyValid, isApiKeyRequired } =
    useApiKeyHandling();
  const { setApiKey: storeApiKey } = useApiActions();
  const apiKeyFromLocalStorage = useApiKey();

  useRefreshSession({
    enabled: process.env.VITE?.VITE_IS_ENTERPRISE === "true",
  });

  /**
   * clean up old api keys from local storage
   *
   * this must be in a useEffect, otherwise we get a warning about changing state during render
   * https://github.com/facebook/react/issues/18178
   */
  useEffect(() => {
    // when no key is required, make sure to delete a possibly existing key from local storage
    if (isFetched && !isApiKeyRequired && apiKeyFromLocalStorage) {
      storeApiKey(null);
    }
  }, [apiKeyFromLocalStorage, isApiKeyRequired, isFetched, storeApiKey]);

  // return nothing until we know the status of the api key and server
  if (!isFetched) return null;

  // when the current key is not valid we show the auth dialog
  if (!isCurrentKeyValid) {
    return <Authdialog />;
  }

  return <>{children}</>;
};
