import { FC } from "react";
import { LogoutButtonProps } from "./types";
import { useApiActions } from "~/util/store/apiKey";

const OpenSourceLogoutButton: FC<LogoutButtonProps> = ({
  children,
  button: Button,
}) => {
  const { setApiKey: storeApiKey } = useApiActions();

  const logout = () => {
    storeApiKey(null);
  };

  return <Button onClick={logout}>{children}</Button>;
};

export default OpenSourceLogoutButton;
