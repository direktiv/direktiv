import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";

import { BlockPathType } from "../../PageCompiler/Block";
import Button from "~/design/Button";
import { FC } from "react";
import { useTranslation } from "react-i18next";

type BlockContextMenuProps = {
  path: BlockPathType;
  onDelete: () => void;
};

export const BlockContextMenu: FC<BlockContextMenuProps> = ({ onDelete }) => {
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
        <DropdownMenuItem className="w-full" onClick={onDelete}>
          <Trash className="mr-2 size-4" />
          {t("direktivPage.blockEditor.contextMenu.deleteButton")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
