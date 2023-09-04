import { MousePointerSquare, MousePointerSquareDashed } from "lucide-react";

import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import { useTranslation } from "react-i18next";

const PermissionsSelector = ({
  availablePermissions,
  selectedPermissions,
  setPermissions,
}: {
  availablePermissions: string[];
  selectedPermissions: string[];
  setPermissions: (permissions: string[]) => void;
}) => {
  const { t } = useTranslation();

  const onCheckedChange = (permissionValue: string, isChecked: boolean) => {
    const currentPermissions = selectedPermissions;
    const newPermissions = isChecked
      ? [...currentPermissions, permissionValue]
      : currentPermissions.filter((p) => p !== permissionValue);
    setPermissions(newPermissions);
  };

  const allSelected =
    selectedPermissions.length === availablePermissions?.length;
  const noneSelected = selectedPermissions.length === 0;

  const selectAllPermissions = () => {
    setPermissions(availablePermissions);
  };

  const deselectAllPermissions = () => {
    setPermissions([]);
  };

  return (
    <>
      <fieldset className="flex items-center gap-5">
        <label className="w-[90px] text-right text-[14px]">
          {t("pages.permissions.permissionsSelector.permissions")}
        </label>
        <div className="grid w-full gap-2 sm:grid-cols-3 ">
          {availablePermissions?.map((permission) => (
            <label
              key={permission}
              className="flex items-center gap-2 text-sm"
              htmlFor={permission}
            >
              <Checkbox
                id={permission}
                value={permission}
                checked={selectedPermissions.includes(permission)}
                onCheckedChange={(checked) => {
                  if (checked !== "indeterminate") {
                    onCheckedChange(permission, checked);
                  }
                }}
              />
              {permission}
            </label>
          ))}
        </div>
      </fieldset>
      <div className="flex justify-end gap-x-2">
        <Button
          variant="link"
          type="button"
          size="sm"
          onClick={selectAllPermissions}
          disabled={allSelected}
        >
          <MousePointerSquare />
          {t("pages.permissions.permissionsSelector.selectAll")}
        </Button>
        <Button
          variant="link"
          size="sm"
          type="button"
          onClick={deselectAllPermissions}
          disabled={noneSelected}
        >
          <MousePointerSquareDashed />
          {t("pages.permissions.permissionsSelector.deselectAll")}
        </Button>
      </div>
    </>
  );
};

export default PermissionsSelector;
