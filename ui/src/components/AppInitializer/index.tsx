import { FC, PropsWithChildren } from "react";

import { useVersion } from "~/api/version/query/get";

export const AppInitializer: FC<PropsWithChildren> = ({ children }) => {
  const { isFetched, data } = useVersion();
  if (!isFetched) return null;

  window._direktiv = {
    ...window._direktiv,
    isEnterprise: data?.data.isEnterprise,
  };

  return <>{children}</>;
};
