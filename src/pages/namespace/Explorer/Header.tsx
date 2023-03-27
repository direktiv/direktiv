import * as Dialog from "@radix-ui/react-dialog";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../../componentsNext/Dropdown";
import { FC, Fragment, useState } from "react";
import { Folder, FolderTree, Play, PlusCircle } from "lucide-react";

import Button from "../../../componentsNext/Button";
import { Link } from "react-router-dom";
import NewDirectory from "./NewDirectory";
import { RxChevronDown } from "react-icons/rx";
import { analyzePath } from "../../../util/router/utils";
import clsx from "clsx";
import { pages } from "../../../util/router/pages";
import { useListDirectory } from "../../../api/tree/query/get";
import { useNamespace } from "../../../util/store/namespace";

const BreadcrumbSegment: FC<{
  absolute: string;
  relative: string;
  namespace: string;
}> = ({ absolute, relative, namespace }) => (
  <Link
    to={pages.explorer.createHref({ namespace, path: absolute })}
    className="hover:underline"
  >
    {relative}
  </Link>
);

const ExplorerHeader: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();

  const { data } = useListDirectory({ path });
  const { segments } = analyzePath(path);
  const [dialogOpen, setDialogOpen] = useState(false);

  if (!namespace) return null;
  return (
    <div className="space-y-5 border-b border-gray-5 bg-base-200 p-5 dark:border-gray-dark-5">
      <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
          <FolderTree className="h-5" />
          <div>
            /&nbsp;
            {segments
              .map((x) => (
                <BreadcrumbSegment
                  key={x.absolute}
                  absolute={x.absolute}
                  relative={x.relative}
                  namespace={namespace}
                />
              ))
              // add / between segments
              .reduce((prev, curr, i) => {
                if (i === 0) return [curr];
                return [
                  ...prev,
                  <Fragment key={`${curr.key}-separator`}> / </Fragment>,
                  curr,
                ];
              }, [] as JSX.Element[])}
          </div>
        </h3>
        <Dialog.Root open={dialogOpen} onOpenChange={setDialogOpen}>
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
                  <Folder className="mr-2 h-4 w-4" />
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
                  "fixed z-50 grid w-full gap-2 rounded-b-lg bg-base-100 p-6 shadow-md animate-in data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 sm:max-w-[425px] sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0"
                )}
              >
                <NewDirectory
                  path={data?.node?.path}
                  unallowedNames={(data?.children?.results ?? []).map(
                    (x) => x.name
                  )}
                  close={() => setDialogOpen(false)}
                />
              </Dialog.Content>
            </div>
          </Dialog.Portal>
        </Dialog.Root>
      </div>
    </div>
  );
};

export default ExplorerHeader;
