import { User, WebStorageStateStore } from "oidc-client-ts";

import { AuthProviderProps } from "react-oidc-context";
import { isDev } from "~/config/env/utils";

const isBrowser = typeof window !== "undefined";

const rootUrl = isBrowser
  ? `${window.location.protocol}//${window.location.host}/`
  : "/";

const client_id = "direktiv";

const getAuthority = () =>
  isDev() ? `${rootUrl}dex` : `${rootUrl}auth/realms/direktiv`;

export const getOidcConfig = (): AuthProviderProps => ({
  authority: getAuthority(),
  client_id,
  post_logout_redirect_uri: rootUrl,
  redirect_uri: rootUrl,
  scope: isDev() ? "openid profile email groups" : undefined,

  /**
   * removes code and state from url after signin
   * see https://github.com/authts/react-oidc-context/blob/f175dcba6ab09871b027d6a2f2224a17712b67c5/src/AuthProvider.tsx#L20-L30
   */
  onSigninCallback: () => {
    window.history.replaceState({}, document.title, window.location.pathname);
  },
  /**
   * we need to store the user in local storage, to access the token. The alternative would
   * be to read it from the user object returned from useAuth, but as only the enterprise
   * edition uses oidc, we would have to conditionally call the hook, which is not possible.
   */
  userStore: isBrowser
    ? new WebStorageStateStore({
        store: window.localStorage,
      })
    : undefined,
});

export const getOidcUser = () => {
  const oidcStorage = localStorage.getItem(
    `oidc.user:${getAuthority()}:${client_id}`
  );
  if (!oidcStorage) {
    return null;
  }

  return User.fromStorageString(oidcStorage);
};
