import * as Dialog from "@radix-ui/react-dialog";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../../componentsNext/Dropdown";
import {
  Folder,
  FolderTree,
  FolderUp,
  Github,
  Play,
  PlusCircle,
} from "lucide-react";

import Button from "../../../componentsNext/Button";
import { FC } from "react";
import { Link } from "react-router-dom";
import { RxChevronDown } from "react-icons/rx";
import { analyzePath } from "../../../util/router/utils";
import clsx from "clsx";
import moment from "moment";
import { pages } from "../../../util/router/pages";
import { useNamespace } from "../../../util/store/namespace";
import { useTree } from "../../../api/tree";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();

  const { data } = useTree({ path });
  const { parent, isRoot } = analyzePath(path);

  if (!namespace) return null;

  return (
    <div>
      <div className="space-y-5 border-b border-gray-5 bg-base-200 p-5 dark:border-gray-dark-5">
        <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between ">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <FolderTree className="h-5" />
            {data?.node?.path}
          </h3>
          <Dialog.Root>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="primary">
                  <PlusCircle /> New <RxChevronDown />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-40">
                <DropdownMenuLabel>Create</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <Dialog.Trigger>
                  <DropdownMenuItem>
                    <FolderUp className="mr-2 h-4 w-4" />
                    <span>New Directory</span>
                  </DropdownMenuItem>
                </Dialog.Trigger>
                <DropdownMenuItem>
                  <Play className="mr-2 h-4 w-4" />
                  <span>New Workflow</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <Dialog.Portal>
              <div className="fixed inset-0 z-50 flex items-start justify-center sm:items-center">
                <Dialog.Overlay
                  className={clsx(
                    "fixed inset-0 z-50 backdrop-blur-sm transition-all duration-100 data-[state=closed]:animate-out data-[state=open]:fade-in data-[state=closed]:fade-out",
                    "bg-black-alpha-2",
                    "dark:bg-white-alpha-2"
                  )}
                />
                <Dialog.Content
                  className={clsx(
                    "fixed z-50 grid w-full gap-2 rounded-b-lg p-6 animate-in data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 sm:max-w-[425px] sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0",
                    "bg-base-100"
                  )}
                >
                  <Dialog.Title className="text-mauve12 m-0 flex gap-2 text-[17px] font-medium">
                    <Folder /> Create a new Folder
                  </Dialog.Title>
                  <Dialog.Description className="text-mauve11 mt-[10px] mb-5 text-[15px] leading-normal">
                    Please enter the name of the new folder.
                  </Dialog.Description>
                  <fieldset className="mb-[15px] flex items-center gap-5">
                    <label
                      className="text-violet11 w-[90px] text-right text-[15px]"
                      htmlFor="name"
                    >
                      Name
                    </label>
                    <input
                      className="text-violet11 shadow-violet7 focus:shadow-violet8 inline-flex h-[35px] w-full flex-1 items-center justify-center rounded-[4px] px-[10px] text-[15px] leading-none shadow-[0_0_0_1px] outline-none focus:shadow-[0_0_0_2px]"
                      id="name"
                      defaultValue="Folder Name"
                    />
                  </fieldset>
                  <div className="flex justify-end gap-2">
                    <Dialog.Close asChild>
                      <Button variant="ghost">Cancel</Button>
                    </Dialog.Close>
                    <Dialog.Close asChild>
                      <Button>Create</Button>
                    </Dialog.Close>
                  </div>
                </Dialog.Content>
              </div>
            </Dialog.Portal>
          </Dialog.Root>
        </div>
      </div>

      <div className="flex flex-col space-y-5 p-5 text-sm">
        <div className="flex flex-col space-y-5 ">
          {!isRoot && (
            <Link
              to={pages.explorer.createHref({
                namespace,
                path: parent?.absolute,
              })}
              className="flex items-center space-x-3"
            >
              <FolderUp className="h-5" />
              <span>..</span>
            </Link>
          )}
          {data?.children?.results.map((file) => {
            let Icon = Folder;
            if (file.expandedType === "workflow") {
              Icon = Play;
            }
            if (file.expandedType === "git") {
              Icon = Github;
            }

            const linkTarget =
              file.expandedType === "workflow"
                ? pages.workflow.createHref({
                    namespace,
                    path: file.path,
                  })
                : pages.explorer.createHref({
                    namespace,
                    path: file.path,
                  });

            return (
              <div key={file.name}>
                <Link to={linkTarget} className="flex items-center space-x-3">
                  <Icon className="h-5" />
                  <span className="flex-1">{file.name}</span>
                  <span className="text-gray-8 dark:text-gray-dark-8">
                    {moment(file.updatedAt).fromNow()}
                  </span>
                </Link>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default ExplorerPage;
