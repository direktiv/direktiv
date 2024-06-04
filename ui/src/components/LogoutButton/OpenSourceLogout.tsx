import { FC } from "react";
import { LogoutButtonProps } from "./types";
import { useApiActions } from "~/util/store/apiKey";

const OpenSourceLogoutButton: FC<LogoutButtonProps> = ({
  children,
  wrapper: Wrapper,
}) => {
  const { setApiKey: storeApiKey } = useApiActions();

  const logout = () => {
    storeApiKey(null);
  };

  return <Wrapper onClick={logout}>{children}</Wrapper>;
};

export default OpenSourceLogoutButton;
