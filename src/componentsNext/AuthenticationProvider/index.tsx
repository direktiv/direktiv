import { FC, PropsWithChildren } from "react";

import { Authdialog } from "../Authdialog";
import useApiKeyHandling from "~/hooksNext/useApiKeyHandling";

export const AuthenticationProvider: FC<PropsWithChildren> = ({ children }) => {
  const { isSuccess, isCurrentKeyValid } = useApiKeyHandling();

  if (!isSuccess) return null;

  if (!isCurrentKeyValid) {
    return <Authdialog />;
  }

  return <>{children}</>;
};
