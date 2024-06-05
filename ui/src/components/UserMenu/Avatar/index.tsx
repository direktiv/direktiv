import AvatarDesignComponent from "~/design/Avatar";
import EnterpriseAvatar from "./EnterpriseAvatar";
import OpenSourceAvatar from "./OpenSourceAvatar";
import { isEnterprise } from "~/config/env/utils";

const Avatar = () => (
  <AvatarDesignComponent>
    {isEnterprise() ? <EnterpriseAvatar /> : <OpenSourceAvatar />}
  </AvatarDesignComponent>
);

export default Avatar;
