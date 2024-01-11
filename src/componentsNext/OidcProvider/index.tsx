import { AuthProvider, AuthProviderProps } from "react-oidc-context";
import { FC, PropsWithChildren } from "react";

import { OidcHandler } from "./OidcHandler";

const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;
const rootUrl = `${window.location.protocol}//${window.location.host}`;
const appFolder = process.env.VITE?.VITE_BASE ?? "/";

const oidcConfig: AuthProviderProps = {
  authority: `${rootUrl}/realms/direktiv`,
  resource: "direktiv-ui",
  client_id: "direktiv-ui",
  redirect_uri: `${rootUrl}${appFolder}`,
  /**
   * removes code and state from url after signin
   * see https://github.com/authts/react-oidc-context/blob/f175dcba6ab09871b027d6a2f2224a17712b67c5/src/AuthProvider.tsx#L20-L30
   */
  onSigninCallback: () => {
    window.history.replaceState({}, document.title, window.location.pathname);
  },
};

export const OidcProvider: FC<PropsWithChildren> = ({ children }) => {
  if (!isEnterprise) {
    return <>{children}</>;
  }
  return (
    <AuthProvider {...oidcConfig}>
      <OidcHandler>{children}</OidcHandler>
    </AuthProvider>
  );
};
