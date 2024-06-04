import { FC } from "react";
import { LogoutButtonProps } from "./types";
import { useAuth } from "react-oidc-context";

const EnterpriseLogoutButton: FC<LogoutButtonProps> = ({
  children,
  wrapper: Wrapper,
}) => {
  const auth = useAuth();

  const logout = () => {
    auth.signoutRedirect();
  };

  return <Wrapper onClick={logout}>{children}</Wrapper>;
};

export default EnterpriseLogoutButton;
