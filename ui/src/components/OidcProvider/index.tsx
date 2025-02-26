import { FC, PropsWithChildren } from "react";

import { AuthProvider } from "react-oidc-context";
import { OidcHandler } from "./OidcHandler";
import { getOidcConfig } from "./utils";
import { isEnterprise } from "~/config/env/utils";

export const OidcProvider: FC<PropsWithChildren> = ({ children }) => {
  if (!isEnterprise()) {
    return <>{children}</>;
  }
  return (
    <AuthProvider {...getOidcConfig()}>
      <OidcHandler>{children}</OidcHandler>
    </AuthProvider>
  );
};
