import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

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
    <HoverCard>
      <HoverCardTrigger>
        <Badge className="cursor-pointer" variant={badgeVariant}>
          {hasAllPermissions
            ? t("pages.permissions.permissionsInfo.all")
            : t("pages.permissions.permissionsInfo.partial", {
                count: permissions.length,
              })}
        </Badge>
      </HoverCardTrigger>
      <HoverCardContent className="flex max-w-xl flex-wrap gap-3">
        {availablePermissions.map((permission) => (
          <Badge
            key={permission}
            className="cursor-pointer"
            variant={permissions.includes(permission) ? "success" : "outline"}
          >
            {permission}
          </Badge>
        ))}
      </HoverCardContent>
    </HoverCard>
  );
};

export default PermissionsInfo;
