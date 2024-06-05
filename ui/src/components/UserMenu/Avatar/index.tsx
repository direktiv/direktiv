import AvatarCompo from "~/design/Avatar";
import EnterpriseAvatar from "./EnterpriseAvatar";
import OpenSourceAvatar from "./OpenSourceAvatar";
import { isEnterprise } from "~/config/env/utils";

const Avatar = () => (
  <AvatarCompo>
    {isEnterprise() ? <EnterpriseAvatar /> : <OpenSourceAvatar />}
  </AvatarCompo>
);

export default Avatar;
