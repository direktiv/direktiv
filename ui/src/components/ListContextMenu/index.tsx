import { ArrowDown, ArrowUp, MoreVertical, Trash } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";

import Button from "~/design/Button";
import { DialogTrigger } from "@radix-ui/react-dialog";
import { FC } from "react";
import { useTranslation } from "react-i18next";

type ListContextMenuProps = {
  onDelete: () => void;
  onMoveUp?: () => void;
  onMoveDown?: () => void;
};

export const ListContextMenu: FC<ListContextMenuProps> = ({
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
              {t("components.contextMenu.moveDownBtn")}
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
              {t("components.contextMenu.moveUpBtn")}
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
            {t("components.contextMenu.deleteBtn")}
          </DropdownMenuItem>
        </DialogTrigger>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
