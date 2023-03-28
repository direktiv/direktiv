import * as Dialog from "@radix-ui/react-dialog";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../../design/Dropdown";
import {
  Folder,
  FolderUp,
  Github,
  MoreVertical,
  Play,
  TextCursorInput,
  Trash,
} from "lucide-react";

import Button from "../../../design/Button";
import Delete from "./Delete";
import ExplorerHeader from "./Header";
import { FC } from "react";
import { Link } from "react-router-dom";
import { analyzePath } from "../../../util/router/utils";
import clsx from "clsx";
import moment from "moment";
import { pages } from "../../../util/router/pages";
import { useListDirectory } from "../../../api/tree/query/get";
import { useNamespace } from "../../../util/store/namespace";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();

  const { data } = useListDirectory({ path });
  const { parent, isRoot } = analyzePath(path);

  if (!namespace) return null;

  return (
    <div>
      <ExplorerHeader />
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
                <div className="flex items-center space-x-3">
                  <Icon className="h-5" />
                  <Link to={linkTarget} className="flex flex-1">
                    <span className="flex-1">{file.name}</span>
                    <span className="text-gray-8 dark:text-gray-dark-8">
                      {moment(file.updatedAt).fromNow()}
                    </span>
                  </Link>
                  <Dialog.Root>
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
                        <DropdownMenuLabel>Edit</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <Dialog.Trigger>
                          <DropdownMenuItem>
                            <Trash className="mr-2 h-4 w-4" />
                            <span>Delete</span>
                          </DropdownMenuItem>
                        </Dialog.Trigger>
                        <Dialog.Trigger>
                          <DropdownMenuItem>
                            <TextCursorInput className="mr-2 h-4 w-4" />
                            <span>Rename</span>
                          </DropdownMenuItem>
                        </Dialog.Trigger>
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
                            "fixed z-50 grid w-full gap-2 rounded-b-lg bg-base-100 p-6 shadow-md animate-in data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 sm:max-w-[425px] sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0"
                          )}
                        >
                          <Delete node={file} close={() => null} />
                        </Dialog.Content>
                      </div>
                    </Dialog.Portal>
                  </Dialog.Root>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default ExplorerPage;
