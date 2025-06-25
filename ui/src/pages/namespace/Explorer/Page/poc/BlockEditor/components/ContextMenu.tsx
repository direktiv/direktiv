import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { Edit, MoreVertical, Trash } from "lucide-react";

import Button from "~/design/Button";
import { FC } from "react";
import { useTranslation } from "react-i18next";

type BlockContextMenuProps = {
  onDelete: () => void;
  onEdit: () => void;
};

export const BlockContextMenu: FC<BlockContextMenuProps> = ({
  onDelete,
  onEdit,
}) => {
  const { t } = useTranslation();
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          className="border-2 border-solid border-gray-8 bg-white dark:border-gray-10 dark:bg-black"
          size="sm"
          onClick={(e) => e.preventDefault()}
          icon
        >
          <MoreVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-40">
        <DropdownMenuItem className="w-full" onClick={onEdit}>
          <Edit className="mr-2 size-4" />
          {t("direktivPage.blockEditor.contextMenu.editButton")}
        </DropdownMenuItem>

        <DropdownMenuItem className="w-full" onClick={onDelete}>
          <Trash className="mr-2 size-4" />
          {t("direktivPage.blockEditor.contextMenu.deleteButton")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
