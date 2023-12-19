import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import {
  NoPermissions,
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import Delete from "./components/modals/Delete";
import ExplorerHeader from "./Header";
import FilePicker from "~/componentsNext/FilePicker";
import FileRow from "./FileRow";
import FileViewer from "./components/modals/FileViewer";
import { FolderUp } from "lucide-react";
import { Link } from "react-router-dom";
import NoResult from "./NoResult";
import { NodeSchemaType } from "~/api/tree/schema/node";
import Rename from "./components/modals/Rename";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();
  const { data, isSuccess, isFetched, isAllowed, noPermissionMessage } =
    useNodeContent({
      path,
    });
  const { parent, isRoot } = analyzePath(path);
  const [dialogOpen, setDialogOpen] = useState(false);

  // we only want to use one dialog component for the whole list,
  // so when the user clicks on the delete button in the list, we
  // set the pointer to that node for the dialog
  const [deleteNode, setDeleteNode] = useState<NodeSchemaType>();
  const [renameNode, setRenameNode] = useState<NodeSchemaType>();
  const [previewNode, setPreviewNode] = useState<NodeSchemaType>();
  const { t } = useTranslation();

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteNode(undefined);
      setRenameNode(undefined);
      setPreviewNode(undefined);
    }
  }, [dialogOpen]);

  if (!namespace) return null;
  if (!isFetched) return null;

  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  const results = data?.children?.results ?? [];
  const showTable = !isRoot || results.length > 0;
  const noResults = isSuccess && results.length === 0;
  const wideOverlay = !!previewNode;

  return (
    <>
      <ExplorerHeader />
      <div className="p-5">
        <Card>
          {showTable && (
            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
              <Table>
                <TableBody>
                  {!isRoot && (
                    <TableRow>
                      <TableCell colSpan={2}>
                        <Link
                          to={pages.explorer.createHref({
                            namespace,
                            path: parent?.absolute,
                          })}
                          className="flex items-center space-x-3 hover:underline"
                        >
                          <FolderUp className="h-5" />
                          <span>
                            {t("pages.explorer.tree.list.oneLevelUp")}
                          </span>
                        </Link>
                      </TableCell>
                    </TableRow>
                  )}
                  {results.map((file) => (
                    <FileRow
                      key={file.name}
                      namespace={namespace}
                      node={file}
                      onDeleteClicked={setDeleteNode}
                      onRenameClicked={setRenameNode}
                      onPreviewClicked={setPreviewNode}
                    />
                  ))}
                </TableBody>
              </Table>
              <DialogContent
                className={twMergeClsx(
                  wideOverlay && "sm:max-w-xl md:max-w-2xl lg:max-w-3xl"
                )}
              >
                {previewNode && <FileViewer node={previewNode} />}
                {deleteNode && (
                  <Delete
                    node={deleteNode}
                    close={() => {
                      setDialogOpen(false);
                    }}
                  />
                )}
                {renameNode && (
                  <Rename
                    node={renameNode}
                    close={() => {
                      setDialogOpen(false);
                    }}
                    unallowedNames={
                      data?.children?.results.map((file) => file.name) || []
                    }
                  />
                )}
              </DialogContent>
            </Dialog>
          )}
          {noResults && <NoResult />}
        </Card>
      </div>
    </>
  );
};

export default ExplorerPage;
