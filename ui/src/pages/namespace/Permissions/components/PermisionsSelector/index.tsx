import {
  PermisionSchemaType,
  permissionMethodsAvailableUi,
  permissionTopics,
} from "~/api/enterprise/schema";
import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { setPermissionForAllTopics, updatePermissions } from "./utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { PermissionRow } from "./Row";
import { SquareMousePointer } from "lucide-react";
import { useTranslation } from "react-i18next";

type PermisionsSelectorProps = {
  permissions: PermisionSchemaType[];
  onChange: (permissions: PermisionSchemaType[]) => void;
};

const PermissionsSelector = ({
  permissions,
  onChange,
}: PermisionsSelectorProps) => {
  const { t } = useTranslation();
  return (
    <fieldset className="flex flex-col items-center gap-5">
      <label className="w-[120px] text-right text-[14px]">
        {t("pages.permissions.permissionsSelector.permissions")}
      </label>
      <Card className="max-h-[400px] w-full overflow-scroll" noShadow>
        <Table>
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell sticky className="px-0"></TableHeaderCell>
              <TableHeaderCell sticky className="w-36 px-2 text-center">
                <div className="flex flex-col gap-2">
                  {t("pages.permissions.permissionsSelector.noPermissions")}
                  <Button
                    variant="outline"
                    size="sm"
                    type="button"
                    disabled={permissions.length === 0}
                    onClick={() => {
                      onChange([]);
                    }}
                  >
                    <SquareMousePointer />
                    {t("pages.permissions.permissionsSelector.selectAll")}
                  </Button>
                </div>
              </TableHeaderCell>
              {permissionMethodsAvailableUi.map((method) => {
                const allSelected =
                  permissions.length > 0 &&
                  permissions.every((p) => p.method === method);
                return (
                  <TableHeaderCell
                    sticky
                    key={method}
                    className="w-36 px-2 text-center"
                  >
                    <div className="flex flex-col gap-2">
                      {method}
                      <Button
                        variant="outline"
                        size="sm"
                        type="button"
                        disabled={allSelected}
                        onClick={() => {
                          onChange(setPermissionForAllTopics(method));
                        }}
                      >
                        <SquareMousePointer />
                        {t("pages.permissions.permissionsSelector.selectAll")}
                      </Button>
                    </div>
                  </TableHeaderCell>
                );
              })}
            </TableRow>
          </TableHead>
          <TableBody>
            {permissionTopics.map((topic) => (
              <PermissionRow
                key={topic}
                topic={topic}
                defaultValue={
                  permissions.find((p) => p.topic === topic)?.method
                }
                onChange={(value) => {
                  onChange(updatePermissions({ permissions, topic, value }));
                }}
              />
            ))}
          </TableBody>
        </Table>
      </Card>
    </fieldset>
  );
};

export default PermissionsSelector;
