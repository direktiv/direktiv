import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { Folder, Layers, Network, Play, PlusCircle } from "lucide-react";

import Button from "~/design/Button";
import { DialogTrigger } from "@radix-ui/react-dialog";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";
import { useTranslation } from "react-i18next";

export type FileTypeSelection =
  | "new-dir"
  | "new-workflow"
  | "new-service"
  | "new-endpoint";

type NewFileButtonProps = {
  setSelectedDialog: (fileType: FileTypeSelection) => void;
};

const NewFileButton: FC<NewFileButtonProps> = ({ setSelectedDialog }) => {
  const { t } = useTranslation();
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="primary" data-testid="dropdown-trg-new">
          <PlusCircle />
          {t("pages.explorer.tree.header.newBtn")}
          <RxChevronDown />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-40">
        <DropdownMenuLabel>
          {t("pages.explorer.tree.header.createLabel")}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DialogTrigger
            className="w-full"
            data-testid="new-dir"
            onClick={() => {
              setSelectedDialog("new-dir");
            }}
          >
            <DropdownMenuItem>
              <Folder className="mr-2 h-4 w-4" />{" "}
              {t("pages.explorer.tree.header.newDirectory")}
            </DropdownMenuItem>
          </DialogTrigger>
          <DialogTrigger
            className="w-full"
            data-testid="new-workflow"
            onClick={() => {
              setSelectedDialog("new-workflow");
            }}
          >
            <DropdownMenuItem>
              <Play className="mr-2 h-4 w-4" />{" "}
              {t("pages.explorer.tree.header.newWorkflow")}
            </DropdownMenuItem>
          </DialogTrigger>
          <DialogTrigger
            className="w-full"
            onClick={() => {
              setSelectedDialog("new-service");
            }}
          >
            <DropdownMenuItem>
              <Layers className="mr-2 h-4 w-4" />{" "}
              {t("pages.explorer.tree.header.newService")}
            </DropdownMenuItem>
          </DialogTrigger>

          <DialogTrigger
            className="w-full"
            onClick={() => {
              setSelectedDialog("new-endpoint");
            }}
          >
            <DropdownMenuItem>
              <Network className="mr-2 h-4 w-4" />{" "}
              {t("pages.explorer.tree.header.newEndpoint")}
            </DropdownMenuItem>
          </DialogTrigger>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export default NewFileButton;
