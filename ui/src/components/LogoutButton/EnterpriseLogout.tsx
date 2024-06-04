import { FC } from "react";
import { LogoutButtonProps } from "./types";
import { useAuth } from "react-oidc-context";

const EnterpriseLogoutButton: FC<LogoutButtonProps> = ({
  children,
  button: Button,
}) => {
  const auth = useAuth();

  const logout = () => {
    auth.signoutRedirect();
  };

  return <Button onClick={logout}>{children}</Button>;
};

export default EnterpriseLogoutButton;
