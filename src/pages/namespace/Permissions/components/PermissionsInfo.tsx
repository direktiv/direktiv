import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { usePermissionKeys } from "~/api/enterprise/permissions/query/get";
import usePermissionLevel from "./usePermissionLevel";
import { useTranslation } from "react-i18next";

const PermissionsInfo = ({ permissions }: { permissions: string[] }) => {
  const { t } = useTranslation();
  const { data: availablePermissions, isSuccess } = usePermissionKeys();
  const { hasAllPermissions, badgeVariant } = usePermissionLevel(
    permissions,
    availablePermissions ?? []
  );

  if (!isSuccess) return null;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <Badge className="cursor-pointer" variant={badgeVariant}>
            {hasAllPermissions
              ? t("pages.permissions.permissionsInfo.all")
              : t("pages.permissions.permissionsInfo.partial", {
                  count: permissions.length,
                })}
          </Badge>
        </TooltipTrigger>
        <TooltipContent className="flex max-w-xl flex-wrap gap-3">
          {availablePermissions.map((permission) => (
            <Badge
              key={permission}
              className="cursor-pointer"
              variant={permissions.includes(permission) ? "success" : "outline"}
            >
              {permission}
            </Badge>
          ))}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};

export default PermissionsInfo;
