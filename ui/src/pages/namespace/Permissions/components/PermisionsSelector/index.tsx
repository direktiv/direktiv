import { MousePointerSquare, MousePointerSquareDashed } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import {
  groupPermissionStringsByResouce,
  joinPermissionString,
  permissionStringsToScopes,
} from "./utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Checkbox } from "~/design/Checkbox";
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

  const availableScopes = useMemo(
    () => permissionStringsToScopes(availablePermissions),
    [availablePermissions]
  );

  const resourceGroups = useMemo(
    () => groupPermissionStringsByResouce(availablePermissions),
    [availablePermissions]
  );

  return (
    <>
      <fieldset className="flex items-center gap-5">
        <label className="w-[90px] text-right text-[14px]">
          {t("pages.permissions.permissionsSelector.permissions")}
        </label>

        <Card
          className="flex w-full flex-col gap-5 overflow-scroll p-5"
          noShadow
        >
          <Table>
            <TableHead>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit lg:pr-8 xl:pr-12">
                <TableHeaderCell />
                {availableScopes.map((scope) => (
                  <TableHeaderCell
                    key={scope}
                    className="w-20 px-2 text-center"
                  >
                    {scope.toLowerCase()}
                  </TableHeaderCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {Object.entries(resourceGroups).map(([resource, scopes]) => (
                <TableRow key={resource} className="lg:pr-8 xl:pr-12">
                  <TableCell className="grow">{resource}</TableCell>
                  {availableScopes.map((availableScope) => {
                    const permissionString = joinPermissionString(
                      availableScope,
                      resource
                    );
                    return (
                      <TableCell
                        key={availableScope}
                        className="px-2 text-center"
                      >
                        {scopes.includes(availableScope) && (
                          <Checkbox
                            checked={selectedPermissions.includes(
                              permissionString
                            )}
                            onCheckedChange={(checked) => {
                              if (checked !== "indeterminate") {
                                onCheckedChange(permissionString, checked);
                              }
                            }}
                          />
                        )}
                      </TableCell>
                    );
                  })}
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <div className="flex justify-end gap-x-2 ">
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
        </Card>
      </fieldset>
    </>
  );
};

export default PermissionsSelector;
