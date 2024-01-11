import { FC, PropsWithChildren } from "react";

import { AuthProvider } from "react-oidc-context";

const oidcConfig = {
  authority: "<your authority>",
  client_id: "<your client id>",
  redirect_uri: "<your redirect uri>",
};

const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;

export const OidcProvider: FC<PropsWithChildren> = ({ children }) => {
  if (!isEnterprise) {
    return <>{children}</>;
  }
  return <AuthProvider {...oidcConfig}>{children}</AuthProvider>;
};
