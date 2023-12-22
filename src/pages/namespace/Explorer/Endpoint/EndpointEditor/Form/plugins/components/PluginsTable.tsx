import { ArrowDown, ArrowUp, MoreVertical, Trash } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, PropsWithChildren } from "react";
import { TableHead, TableHeaderCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "@radix-ui/react-dialog";
import { useTranslation } from "react-i18next";

type TableHeaderProps = PropsWithChildren & {
  title: string;
};

export const TableHeader: FC<TableHeaderProps> = ({ title, children }) => (
  <TableHead>
    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
      <TableHeaderCell>{title}</TableHeaderCell>
      <TableHeaderCell className="w-60 text-right">{children}</TableHeaderCell>
    </TableRow>
  </TableHead>
);

type ContextMenuProps = {
  onDelete: () => void;
  onMoveUp?: () => void;
  onMoveDown?: () => void;
};
export const ContextMenu: FC<ContextMenuProps> = ({
  onDelete,
  onMoveDown,
  onMoveUp,
}) => {
  const { t } = useTranslation();
  return (
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
        {onMoveDown && (
          <DialogTrigger
            className="w-full"
            onClick={(e) => {
              e.stopPropagation();
              e.preventDefault();
              onMoveDown();
            }}
          >
            <DropdownMenuItem>
              <ArrowDown className="mr-2 h-4 w-4" />
              {t("pages.explorer.endpoint.editor.form.plugins.moveDownBtn")}
            </DropdownMenuItem>
          </DialogTrigger>
        )}
        {onMoveUp && (
          <DialogTrigger
            className="w-full"
            onClick={(e) => {
              e.stopPropagation();
              e.preventDefault();
              onMoveUp();
            }}
          >
            <DropdownMenuItem>
              <ArrowUp className="mr-2 h-4 w-4" />
              {t("pages.explorer.endpoint.editor.form.plugins.moveUpBtn")}
            </DropdownMenuItem>
          </DialogTrigger>
        )}
        <DialogTrigger
          className="w-full"
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
            onDelete();
          }}
        >
          <DropdownMenuItem>
            <Trash className="mr-2 h-4 w-4" />
            {t("pages.explorer.endpoint.editor.form.plugins.deleteBtn")}
          </DropdownMenuItem>
        </DialogTrigger>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
