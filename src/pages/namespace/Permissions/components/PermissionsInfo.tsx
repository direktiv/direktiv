import { usePermissionKeys } from "~/api/enterprise/permissions/query/get";
import { useTranslation } from "react-i18next";

const PermissionsInfo = ({ permissions }: { permissions: string[] }) => {
  const { t } = useTranslation();
  const { data: permissionKeys, isSuccess } = usePermissionKeys();
  if (!isSuccess) return null;

  const hasAllPermissions =
    permissionKeys?.length > 0 &&
    permissionKeys?.every((perm) => permissions.includes(perm));

  return (
    <div>
      {hasAllPermissions
        ? t("pages.permissions.permissionsInfo.all")
        : t("pages.permissions.permissionsInfo.partial", {
            count: permissions.length,
          })}
    </div>
  );
};

export default PermissionsInfo;
