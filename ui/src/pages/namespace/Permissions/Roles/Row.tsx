import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Pencil, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import PermissionsInfo from "../components/PermissionsInfo";
import { RoleSchemaType } from "~/api/enterprise/roles/schema";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Row = ({
  role,
  onDeleteClicked,
  onEditClicked,
}: {
  role: RoleSchemaType;
  onDeleteClicked: (group: RoleSchemaType) => void;
  onEditClicked: (group: RoleSchemaType) => void;
}) => {
  const { t } = useTranslation();
  const createdAt = useUpdatedAt(role.createdAt);
  return (
    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
      <TableCell>{role.name}</TableCell>
      <TableCell>{role.description}</TableCell>
      <TableCell>{role.oidcGroups.join(", ")}</TableCell>
      <TableCell>
        <PermissionsInfo permissions={role.permissions} />
      </TableCell>
      <TableCell>{createdAt}</TableCell>
      <TableCell>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => e.preventDefault()}
              icon
            >
              <MoreVertical />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-40">
            <DialogTrigger
              className="w-full"
              onClick={(e) => {
                e.stopPropagation();
                onDeleteClicked(role);
              }}
            >
              <DropdownMenuItem>
                <Trash className="mr-2 size-4" />
                {t("pages.permissions.roles.contextMenu.delete")}
              </DropdownMenuItem>
            </DialogTrigger>
            <DialogTrigger
              className="w-full"
              onClick={(e) => {
                e.stopPropagation();
                onEditClicked(role);
              }}
            >
              <DropdownMenuItem>
                <Pencil className="mr-2 size-4" />
                {t("pages.permissions.roles.contextMenu.edit")}
              </DropdownMenuItem>
            </DialogTrigger>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
};

export default Row;
