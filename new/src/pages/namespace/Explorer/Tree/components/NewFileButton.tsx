import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import {
  Folder,
  Layers,
  Network,
  Play,
  PlusCircle,
  Users,
  Workflow,
} from "lucide-react";

import Button from "~/design/Button";
import { DialogTrigger } from "@radix-ui/react-dialog";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";
import { useTranslation } from "react-i18next";

export type FileTypeSelection =
  | "new-dir"
  | "new-workflow"
  | "new-service"
  | "new-route"
  | "new-consumer";

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
          {t("pages.explorer.tree.newFileButton.buttonText")}
          <RxChevronDown />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-40">
        <DropdownMenuLabel>
          {t("pages.explorer.tree.newFileButton.label")}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DialogTrigger
          className="w-full"
          data-testid="new-dir"
          onClick={() => {
            setSelectedDialog("new-dir");
          }}
        >
          <DropdownMenuItem>
            <Folder className="mr-2 h-4 w-4" />{" "}
            {t("pages.explorer.tree.newFileButton.items.directory")}
          </DropdownMenuItem>
        </DialogTrigger>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DialogTrigger
            className="w-full"
            data-testid="new-workflow"
            onClick={() => {
              setSelectedDialog("new-workflow");
            }}
          >
            <DropdownMenuItem>
              <Play className="mr-2 h-4 w-4" />{" "}
              {t("pages.explorer.tree.newFileButton.items.workflow")}
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
              {t("pages.explorer.tree.newFileButton.items.service")}
            </DropdownMenuItem>
          </DialogTrigger>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              <Network className="mr-2 h-4 w-4" />
              {t("pages.explorer.tree.newFileButton.items.gateway.label")}
            </DropdownMenuSubTrigger>
            <DropdownMenuPortal>
              <DropdownMenuSubContent className="w-40">
                <DialogTrigger
                  className="w-full"
                  onClick={() => {
                    setSelectedDialog("new-route");
                  }}
                >
                  <DropdownMenuItem>
                    <Workflow className="mr-2 h-4 w-4" />
                    {t("pages.explorer.tree.newFileButton.items.gateway.route")}
                  </DropdownMenuItem>
                </DialogTrigger>
                <DialogTrigger
                  className="w-full"
                  onClick={() => {
                    setSelectedDialog("new-consumer");
                  }}
                >
                  <DropdownMenuItem>
                    <Users className="mr-2 h-4 w-4" />
                    {t(
                      "pages.explorer.tree.newFileButton.items.gateway.consumer"
                    )}
                  </DropdownMenuItem>
                </DialogTrigger>
              </DropdownMenuSubContent>
            </DropdownMenuPortal>
          </DropdownMenuSub>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export default NewFileButton;
