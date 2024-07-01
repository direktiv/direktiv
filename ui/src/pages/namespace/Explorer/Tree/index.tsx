import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, useEffect, useMemo, useState } from "react";
import {
  NoPermissions,
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "~/design/Table";

import { BaseFileSchemaType } from "~/api/files/schema";
import { Card } from "~/design/Card";
import Delete from "./components/modals/Delete";
import { DropdownMenuSeparator } from "~/design/Dropdown";
import ExplorerHeader from "./Header";
import FileRow from "./FileRow";
import FileViewer from "./components/modals/FileViewer";
import { FolderUp } from "lucide-react";
import Input from "~/design/Input";
import { Link } from "react-router-dom";
import NoResult from "./NoResult";
import NoSearchResult from "./NoSearchResult";
import Rename from "./components/modals/Rename";
import { analyzePath } from "~/util/router/utils";
import { getFilenameFromPath } from "~/api/files/utils";
import { twMergeClsx } from "~/util/helpers";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const ExplorerPage: FC = () => {
  const pages = usePages();
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

  const [query, setQuery] = useState("");
  const { t } = useTranslation();
  const hasQuery = query.length > 0;

  const children = useMemo(
    () => (data?.type === "directory" && data?.children) || [],
    [data]
  );

  const filteredFiles = useMemo(
    () =>
      children.filter(
        (child) =>
          !hasQuery ||
          getFilenameFromPath(child.path.toLowerCase()).includes(
            query.toLowerCase()
          )
      ),
    [hasQuery, query, children]
  );

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

  const showTable = !isRoot || children.length > 0;
  const noResults = isSuccess && children.length === 0;
  const noSearchResult = isSuccess && filteredFiles.length === 0;
  const wideOverlay = !!previewNode;

  const existingNames = children?.map((file) => getFilenameFromPath(file.path));

  return (
    <>
      <ExplorerHeader />
      <div className="p-5">
        <Card>
          {showTable && (
            <>
              <div className="flex justify-between gap-5 p-2">
                <Input
                  data-testid="queryField"
                  className="sm:w-60"
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value);
                  }}
                  placeholder={t("pages.explorer.tree.list.filter")}
                />
              </div>
              <DropdownMenuSeparator />

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
                    {filteredFiles.map((item) => (
                      <FileRow
                        key={item.path}
                        namespace={namespace}
                        file={item}
                        onDeleteClicked={setDeleteNode}
                        onRenameClicked={setRenameNode}
                        onPreviewClicked={setPreviewNode}
                      />
                    ))}
                  </TableBody>
                </Table>
                {noSearchResult && (
                  <>
                    {!isRoot && <DropdownMenuSeparator />}
                    <NoSearchResult />
                  </>
                )}
                <DialogContent
                  className={twMergeClsx(
                    wideOverlay && "sm:max-w-xl md:max-w-2xl lg:max-w-3xl"
                  )}
                >
                  {previewNode && <FileViewer file={previewNode} />}
                  {deleteNode && (
                    <Delete
                      file={deleteNode}
                      close={() => {
                        setDialogOpen(false);
                      }}
                    />
                  )}
                  {renameFile && (
                    <Rename
                      file={renameFile}
                      close={() => {
                        setDialogOpen(false);
                      }}
                      unallowedNames={existingNames}
                    />
                  )}
                </DialogContent>
              </Dialog>
            </>
          )}
          {noResults && <NoResult />}
        </Card>
      </div>
    </>
  );
};

export default ExplorerPage;
