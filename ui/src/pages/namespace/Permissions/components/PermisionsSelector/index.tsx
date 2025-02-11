import {
  PermisionSchemaType,
  permissionMethodsAvailableUi,
  permissionTopics,
} from "~/api/enterprise/schema";
import { SquareDashedMousePointer, SquareMousePointer } from "lucide-react";
import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { PermissionRow } from "./Row";
import { updatePermissions } from "./utils";
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
                    // TODO:
                    // onClick={selectAllPermissions}
                    // disabled={allSelected}
                  >
                    <SquareMousePointer />
                    {t("pages.permissions.permissionsSelector.selectAll")}
                  </Button>
                  <Button
                    variant="link"
                    size="sm"
                    type="button"
                    // TODO:
                    // onClick={deselectAllPermissions}
                    // disabled={noneSelected}
                  >
                    <SquareDashedMousePointer />
                    {t("pages.permissions.permissionsSelector.deselectAll")}
                  </Button>
                </div>
              </TableHeaderCell>
              <TableHeaderCell sticky className="w-32 px-2 text-center">
                {t("pages.permissions.permissionsSelector.noPermissions")}
              </TableHeaderCell>
              {permissionMethodsAvailableUi.map((method) => (
                <TableHeaderCell
                  sticky
                  key={method}
                  className="w-20 px-2 text-center"
                >
                  {method.toLowerCase()}
                </TableHeaderCell>
              ))}
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
