import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, Fragment, useEffect, useState } from "react";
import { Folder, FolderTree, Play, PlusCircle } from "lucide-react";

import Button from "~/design/Button";
import { DropdownMenuGroup } from "@radix-ui/react-dropdown-menu";
import { Link } from "react-router-dom";
import NewDirectory from "./NewDirectory";
import NewWorkflow from "./NewWorkflow";
import { RxChevronDown } from "react-icons/rx";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const BreadcrumbSegment: FC<{
  absolute: string;
  relative: string;
  namespace: string;
}> = ({ absolute, relative, namespace, ...props }) => (
  <Link
    to={pages.explorer.createHref({ namespace, path: absolute })}
    className="hover:underline"
    {...props}
  >
    {relative}
  </Link>
);

const ExplorerHeader: FC = () => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();

  const { data } = useNodeContent({ path });
  const { segments } = analyzePath(path);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedDialog, setSelectedDialog] = useState<
    "new-dir" | "new-workflow" | undefined
  >();

  useEffect(() => {
    if (dialogOpen === false) setSelectedDialog(undefined);
  }, [dialogOpen, selectedDialog]);

  if (!namespace) return null;
  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
          <Link
            data-testid="tree-root"
            to={pages.explorer.createHref({ namespace })}
            className="hover:underline"
          >
            <FolderTree className="h-5" />
          </Link>
          <div>
            / &nbsp;
            {segments
              .map((x) => (
                <BreadcrumbSegment
                  data-testid="breadcrumb-segment"
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
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
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
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
          <DialogContent>
            {selectedDialog === "new-dir" && (
              <NewDirectory
                path={data?.node?.path}
                unallowedNames={(data?.children?.results ?? []).map(
                  (x) => x.name
                )}
                close={() => setDialogOpen(false)}
              />
            )}
            {selectedDialog === "new-workflow" && (
              <NewWorkflow
                path={data?.node?.path}
                unallowedNames={(data?.children?.results ?? []).map(
                  (file) => file.name
                )}
                close={() => setDialogOpen(false)}
              />
            )}
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

export default ExplorerHeader;
