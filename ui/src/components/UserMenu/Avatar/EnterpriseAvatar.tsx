import { useAuth } from "react-oidc-context";

const EnterpriseAvatar = () => {
  const auth = useAuth();
  const username = auth.user?.profile?.preferred_username ?? "";
  return <>{username.slice(0, 2)}</>;
};

export default EnterpriseAvatar;
