import { User } from "lucide-react";
import { useAuth } from "react-oidc-context";

const EnterpriseAvatar = () => {
  const auth = useAuth();
  const username = auth?.user?.profile?.preferred_username ?? "";

  if (!username) return <User />;

  return <>{username.slice(0, 2)}</>;
};

export default EnterpriseAvatar;
