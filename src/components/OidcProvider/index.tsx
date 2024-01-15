import { FC, PropsWithChildren } from "react";

import { AuthProvider } from "react-oidc-context";
import { OidcHandler } from "./OidcHandler";
import { oidcConfig } from "./utils";

const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;

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
