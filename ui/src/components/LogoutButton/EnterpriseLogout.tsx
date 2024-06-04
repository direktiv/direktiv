import { FC } from "react";
import { LogoutButtonProps } from "./types";
import { useAuth } from "react-oidc-context";

const EnterpriseLogoutButton: FC<LogoutButtonProps> = ({
  children,
  wrapper: WrapperComponent,
}) => {
  const auth = useAuth();

  const logout = () => {
    auth.signoutRedirect();
  };

  return <WrapperComponent onClick={logout}>{children}</WrapperComponent>;
};

export default EnterpriseLogoutButton;
