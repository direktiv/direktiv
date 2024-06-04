import { FC } from "react";
import { LogoutButtonProps } from "./types";
import { useApiActions } from "~/util/store/apiKey";

const OpenSourceLogoutButton: FC<LogoutButtonProps> = ({
  children,
  wrapper: WrapperComponent,
}) => {
  const { setApiKey: storeApiKey } = useApiActions();

  const logout = () => {
    storeApiKey(null);
  };

  return <WrapperComponent onClick={logout}>{children}</WrapperComponent>;
};

export default OpenSourceLogoutButton;
