import { BaseFileSchemaType, getFilenameFromPath } from "~/api/files/schema";
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
import FileRow from "./FileRow";
import FileViewer from "./components/modals/FileViewer";
import { FolderUp } from "lucide-react";
import { Link } from "react-router-dom";
import NoResult from "./NoResult";
import Rename from "./components/modals/Rename";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const ExplorerPage: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();
  const { data, isSuccess, isFetched, isAllowed, noPermissionMessage } =
    useFile({
      path,
    });

  const { parent, isRoot } = analyzePath(path);
  const [dialogOpen, setDialogOpen] = useState(false);

  // we only want to use one dialog component for the whole list,
  // so when the user clicks on the delete button in the list, we
  // set the pointer to that node for the dialog
  const [deleteNode, setDeleteNode] = useState<BaseFileSchemaType>();
  const [renameFile, setRenameNode] = useState<BaseFileSchemaType>();
  const [previewNode, setPreviewNode] = useState<BaseFileSchemaType>();
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

  const children = (data?.type === "directory" && data?.children) || [];
  const showTable = !isRoot || children.length > 0;
  const noResults = isSuccess && children.length === 0;
  const wideOverlay = !!previewNode;

  const existingNames =
    data?.type === "directory" && data.children
      ? data.children?.map((file) => getFilenameFromPath(file.path))
      : [];

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
                  {children.map((item) => (
                    <FileRow
                      key={item.path}
                      namespace={namespace}
                      node={item}
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
                {renameFile && (
                  <Rename
                    node={renameFile}
                    close={() => {
                      setDialogOpen(false);
                    }}
                    unallowedNames={existingNames}
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
