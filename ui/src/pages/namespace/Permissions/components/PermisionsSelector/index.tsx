import { SquareDashedMousePointer, SquareMousePointer } from "lucide-react";
import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import {
  groupPermissionStringsByResource,
  permissionStringsToScopes,
} from "./utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { PermissionRow } from "./Row";
import { useMemo } from "react";
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

  const onCheckedChange = (permission: string, isChecked: boolean) => {
    const currentPermissions = selectedPermissions;
    const newPermissions = isChecked
      ? [...currentPermissions, permission]
      : currentPermissions.filter((p) => p !== permission);
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

  const availableScopes = useMemo(
    () => permissionStringsToScopes(availablePermissions),
    [availablePermissions]
  );

  const groupedResources = useMemo(
    () => groupPermissionStringsByResource(availablePermissions),
    [availablePermissions]
  );

  const sortedResources = useMemo(
    () =>
      Object.entries(groupedResources).sort((a, b) => a[0].localeCompare(b[0])),
    [groupedResources]
  );

  return (
    <>
      <fieldset className="flex items-center gap-5">
        <label className="w-[90px] text-right text-[14px]">
          {t("pages.permissions.permissionsSelector.permissions")}
        </label>

        <Card className="max-h-[400px] w-full overflow-scroll" noShadow>
          <Table>
            <TableHead>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell sticky className="px-0">
                  <div className="flex gap-5">
                    <Button
                      variant="link"
                      type="button"
                      size="sm"
                      onClick={selectAllPermissions}
                      disabled={allSelected}
                    >
                      <SquareMousePointer />
                      {t("pages.permissions.permissionsSelector.selectAll")}
                    </Button>
                    <Button
                      variant="link"
                      size="sm"
                      type="button"
                      onClick={deselectAllPermissions}
                      disabled={noneSelected}
                    >
                      <SquareDashedMousePointer />
                      {t("pages.permissions.permissionsSelector.deselectAll")}
                    </Button>
                  </div>
                </TableHeaderCell>
                {availableScopes.map((scope) => (
                  <TableHeaderCell
                    sticky
                    key={scope}
                    className="w-20 px-2 text-center"
                  >
                    {scope.toLowerCase()}
                  </TableHeaderCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {sortedResources.map(([resource, scopes]) => (
                <PermissionRow
                  key={resource}
                  resource={resource}
                  scopes={scopes}
                  availableScopes={availableScopes}
                  selectedPermissions={selectedPermissions}
                  onCheckedChange={onCheckedChange}
                />
              ))}
            </TableBody>
          </Table>
        </Card>
      </fieldset>
    </>
  );
};

export default PermissionsSelector;
