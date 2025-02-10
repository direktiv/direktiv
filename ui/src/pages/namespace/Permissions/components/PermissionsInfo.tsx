import { PermisionSchemaType, permissionTopics } from "~/api/enterprise/schema";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import usePermissionLevel from "./usePermissionLevel";
import { useTranslation } from "react-i18next";

type PermissionsInfoProps = {
  permissions: PermisionSchemaType[];
};

const PermissionsInfo = ({ permissions }: PermissionsInfoProps) => {
  const { t } = useTranslation();
  const { hasAllPermissions, badgeVariant } = usePermissionLevel(
    permissions.map((permission) => permission.topic),
    [...permissionTopics]
  );

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
        <TooltipContent className="flex flex-col max-w-xl flex-wrap gap-3 text-inherit">
          <Table>
            <TableBody>
              {permissionTopics.map((permission) => {
                const permisionValue = permissions.find(
                  (p) => p.topic === permission
                );

                const permissionSet = !!permisionValue;

                return (
                  <TableRow key={permission}>
                    <TableCell className="font-bold">{permission}</TableCell>
                    <TableCell>
                      {permissionSet
                        ? permisionValue?.method
                        : t("pages.permissions.permissionsInfo.noPermissions")}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};

export default PermissionsInfo;
