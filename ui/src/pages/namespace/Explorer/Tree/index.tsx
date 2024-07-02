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
import EmptyDirectory from "./EmptyDirectory";
import ExplorerHeader from "./Header";
import FileRow from "./FileRow";
import FileViewer from "./components/modals/FileViewer";
import { LevelUpNavigation } from "./LevelUpNavigation";
import NoSearchResult from "./NoSearchResult";
import Rename from "./components/modals/Rename";
import { SearchBar } from "./SearchBar";
import { analyzePath } from "~/util/router/utils";
import { getFilenameFromPath } from "~/api/files/utils";
import { twMergeClsx } from "~/util/helpers";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";

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

  const isEmptyDirectory = isSuccess && children.length === 0;
  const noSearchResult = hasQuery && filteredFiles.length === 0;
  const showEmptyListNote = noSearchResult || isEmptyDirectory;

  const wideOverlay = !!previewNode;
  const existingNames = children?.map((file) => getFilenameFromPath(file.path));

  return (
    <>
      <ExplorerHeader />
      <div className="p-5">
        <Card>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <Table>
              <TableBody>
                {!isEmptyDirectory && (
                  <SearchBar
                    query={query}
                    onChange={(newQuery) => {
                      setQuery(newQuery);
                    }}
                  />
                )}
                {!isRoot && (
                  <LevelUpNavigation
                    namespace={namespace}
                    path={parent?.absolute}
                  />
                )}
                {showEmptyListNote && (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell>
                      {noSearchResult ? <NoSearchResult /> : <EmptyDirectory />}
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
        </Card>
      </div>
    </>
  );
};

export default ExplorerPage;
