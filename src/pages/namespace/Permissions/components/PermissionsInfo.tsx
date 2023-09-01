import Badge from "~/design/Badge";
import { ComponentProps } from "react";
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
    <Badge className="cursor-pointer" variant={badgeVariant}>
      {hasAllPermissions
        ? t("pages.permissions.permissionsInfo.all")
        : t("pages.permissions.permissionsInfo.partial", {
            count: permissions.length,
          })}
    </Badge>
  );
};

export default PermissionsInfo;
