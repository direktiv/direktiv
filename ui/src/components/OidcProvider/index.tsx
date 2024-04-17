import { FC, PropsWithChildren } from "react";

import { AuthProvider } from "react-oidc-context";
import { OidcHandler } from "./OidcHandler";
import { isEnterprise } from "~/config/env/utils";
import { oidcConfig } from "./utils";

export const OidcProvider: FC<PropsWithChildren> = ({ children }) => {
  if (!isEnterprise()) {
    return <>{children}</>;
  }
  return (
    <AuthProvider {...oidcConfig}>
      <OidcHandler>{children}</OidcHandler>
    </AuthProvider>
  );
};
