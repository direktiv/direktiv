import { FC, PropsWithChildren, useEffect, useState } from "react";
import { hasAuthParams, useAuth } from "react-oidc-context";

import Alert from "~/design/Alert";
import { Loader2 } from "lucide-react";
import { getOidcUser } from "./utils";

export const OidcHandler: FC<PropsWithChildren> = ({ children }) => {
  const auth = useAuth();
  const [hasTriedSignin, setHasTriedSignin] = useState(false);

  // TODO: remove this before merging
  // eslint-disable-next-line no-console
  console.log("🔑 oidc auth object", auth);
  // eslint-disable-next-line no-console
  console.log("👤 user", getOidcUser());
  // eslint-disable-next-line no-console
  console.log("⌚ expires at", getOidcUser()?.expires_at);

  useEffect(() => {
    if (
      !hasAuthParams() &&
      !auth.isAuthenticated &&
      !auth.activeNavigator &&
      !auth.isLoading &&
      !hasTriedSignin
    ) {
      auth.signinRedirect();
      setHasTriedSignin(true);
    }
  }, [auth, hasTriedSignin]);

  if (auth.error) {
    return (
      <div className="flex w-full flex-col items-center p-5">
        <Alert variant="error">
          {auth.error.name}: {auth.error.message}
        </Alert>
      </div>
    );
  }

  if (auth.isLoading) {
    return (
      <div className="flex w-full flex-col items-center p-5">
        <Loader2 className="h-4 w-4 animate-spin" />
      </div>
    );
  }

  return <>{children}</>;
};
