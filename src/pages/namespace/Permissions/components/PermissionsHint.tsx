import { BadgeCheck, BadgeHelp } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import CopyButton from "~/design/CopyButton";
import { usePermissionKeys } from "~/api/enterprise/permissions/query/get";
import { useTranslation } from "react-i18next";

const PermissionsHint = () => {
  const { data: availablePermissions } = usePermissionKeys();
  const permissionsAvailable = (availablePermissions ?? []).length > 0;
  const { t } = useTranslation();
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" icon disabled={!permissionsAvailable}>
          <BadgeHelp />
        </Button>
      </PopoverTrigger>
      <PopoverContent asChild>
        <div className="w-max p-5">
          <h3 className="mb-5 flex items-center gap-x-2 font-bold">
            <BadgeCheck className="h-5" />
            {t("pages.permissions.policy.permissionsHintTitle")}
          </h3>
          <div className="grid w-max grid-cols-3 gap-x-8 gap-y-2">
            {availablePermissions?.map((permission) => (
              <div key={permission} className="group flex items-center gap-5">
                <code className="grow text-sm text-primary-500">
                  {permission}
                </code>
                <CopyButton
                  value={permission}
                  buttonProps={{
                    icon: true,
                    variant: "ghost",
                    size: "sm",
                    className: "invisible group-hover:visible",
                  }}
                />
              </div>
            ))}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
};

export default PermissionsHint;
