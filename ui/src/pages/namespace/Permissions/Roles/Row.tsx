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

const Row = ({
  group,
  onDeleteClicked,
  onEditClicked,
}: {
  group: RoleSchemaType;
  onDeleteClicked: (group: RoleSchemaType) => void;
  onEditClicked: (group: RoleSchemaType) => void;
}) => {
  const { t } = useTranslation();
  return (
    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
      <TableCell>{group.group}</TableCell>
      <TableCell>{group.description}</TableCell>
      <TableCell>
        {/* <PermissionsInfo permissions={group.permissions} /> */}
      </TableCell>
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
                onDeleteClicked(group);
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
                onEditClicked(group);
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
