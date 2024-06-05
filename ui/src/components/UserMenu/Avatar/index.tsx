import { ElementRef, forwardRef } from "react";

import AvatarDesignComponent from "~/design/Avatar";
import Button from "~/design/Button";
import EnterpriseAvatar from "./EnterpriseAvatar";
import OpenSourceAvatar from "./OpenSourceAvatar";
import { RxChevronDown } from "react-icons/rx";
import { Settings2 } from "lucide-react";
import { isEnterprise } from "~/config/env/utils";
import useApiKeyHandling from "~/hooks/useApiKeyHandling";

const Avatar = forwardRef<ElementRef<typeof Button>>((props, ref) => {
  const { usesAccounts } = useApiKeyHandling();
  return usesAccounts ? (
    <Button
      ref={ref}
      variant="ghost"
      className="items-center px-1"
      role="button"
      icon
      data-testid="dropdown-trg-user-menu"
      {...props}
    >
      <AvatarDesignComponent>
        {isEnterprise() ? <EnterpriseAvatar /> : <OpenSourceAvatar />}
      </AvatarDesignComponent>
      <RxChevronDown />
    </Button>
  ) : (
    <Button
      ref={ref}
      variant="ghost"
      icon
      data-testid="dropdown-trg-user-menu"
      {...props}
    >
      <Settings2 />
      <RxChevronDown />
    </Button>
  );
});

Avatar.displayName = "Avatar";

export default Avatar;
