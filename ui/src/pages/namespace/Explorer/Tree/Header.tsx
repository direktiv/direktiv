import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, Fragment, useEffect, useState } from "react";
import NewFileButton, { FileTypeSelection } from "./components/NewFileButton";

import { FolderTree } from "lucide-react";
import { Link } from "react-router-dom";
import NewConsumer from "./components/modals/CreateNew/Gateway/Consumer";
import NewDirectory from "./components/modals/CreateNew/Directory";
import NewRoute from "./components/modals/CreateNew/Gateway/Route";
import NewService from "./components/modals/CreateNew/Service";
import NewWorkflow from "./components/modals/CreateNew/Workflow";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";

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
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();

  const { data } = useNodeContent({ path });
  const { segments } = analyzePath(path);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedDialog, setSelectedDialog] = useState<FileTypeSelection>();

  useEffect(() => {
    if (dialogOpen === false) setSelectedDialog(undefined);
  }, [dialogOpen, selectedDialog]);

  if (!namespace) return null;

  const wideOverlay =
    !!selectedDialog &&
    !["new-dir", "new-route", "new-consumer", "new-service"].includes(
      selectedDialog
    );

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
          <NewFileButton setSelectedDialog={setSelectedDialog} />
          <DialogContent
            className={twMergeClsx(
              wideOverlay && "sm:max-w-xl md:max-w-2xl lg:max-w-3xl"
            )}
          >
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
            {selectedDialog === "new-service" && (
              <NewService
                path={data?.node?.path}
                unallowedNames={(data?.children?.results ?? []).map(
                  (x) => x.name
                )}
                close={() => setDialogOpen(false)}
              />
            )}
            {selectedDialog === "new-route" && (
              <NewRoute
                path={data?.node?.path}
                unallowedNames={(data?.children?.results ?? []).map(
                  (x) => x.name
                )}
                close={() => setDialogOpen(false)}
              />
            )}
            {selectedDialog === "new-consumer" && (
              <NewConsumer
                path={data?.node?.path}
                unallowedNames={(data?.children?.results ?? []).map(
                  (x) => x.name
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
