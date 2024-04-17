import { FC, PropsWithChildren } from "react";

import { useVersion } from "~/api/version/query/get";

export const AppInitializer: FC<PropsWithChildren> = ({ children }) => {
  const { isFetched } = useVersion();
  if (!isFetched) return null;

  window._direktiv = {
    ...window._direktiv,
    // isEnterprise: true, // TODO: update the window value
  };

  return <>{children}</>;
};
