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
import { GroupSchemaType } from "~/api/enterprise/groups/schema";
import PermissionsInfo from "../components/PermissionsInfo";
import { useTranslation } from "react-i18next";

const Row = ({
  group,
  onDeleteClicked,
  onEditClicked,
}: {
  group: GroupSchemaType;
  onDeleteClicked: (group: GroupSchemaType) => void;
  onEditClicked: (group: GroupSchemaType) => void;
}) => {
  const { t } = useTranslation();
  return (
    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
      <TableCell>{group.group}</TableCell>
      <TableCell>{group.description}</TableCell>
      <TableCell>
        <PermissionsInfo permissions={group.permissions} />
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
                <Trash className="mr-2 h-4 w-4" />
                {t("pages.permissions.groups.contextMenu.delete")}
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
                <Pencil className="mr-2 h-4 w-4" />
                {t("pages.permissions.groups.contextMenu.edit")}
              </DropdownMenuItem>
            </DialogTrigger>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
};

export default Row;
