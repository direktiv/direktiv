import AvatarDesignComponent from "~/design/Avatar";
import Button from "~/design/Button";
import EnterpriseAvatar from "./EnterpriseAvatar";
import OpenSourceAvatar from "./OpenSourceAvatar";
import { RxChevronDown } from "react-icons/rx";
import { Settings2 } from "lucide-react";
import { isEnterprise } from "~/config/env/utils";
import useApiKeyHandling from "~/hooks/useApiKeyHandling";

const Avatar = () => {
  const { usesAccounts } = useApiKeyHandling();
  return usesAccounts ? (
    <Button
      variant="ghost"
      className="items-center px-1"
      role="button"
      icon
      data-testid="dropdown-trg-user-menu"
    >
      <AvatarDesignComponent>
        {isEnterprise() ? <EnterpriseAvatar /> : <OpenSourceAvatar />}
      </AvatarDesignComponent>
      <RxChevronDown />
    </Button>
  ) : (
    <Button variant="ghost" icon data-testid="dropdown-trg-user-menu">
      <Settings2 />
      <RxChevronDown />
    </Button>
  );
};

export default Avatar;
