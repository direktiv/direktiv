import { NoResult, Table, TableBody } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";
import { useEffect, useMemo, useState } from "react";

import { Braces } from "lucide-react";
import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "~/pages/namespace/Settings/components/CreateItemButton";
import Delete from "~/pages/namespace/Settings/Variables/Delete";
import { Dialog } from "~/design/Dialog";
import Edit from "./Edit";
import Input from "~/design/Input";
import ItemRow from "~/pages/namespace/Settings/components/ItemRow";
import PaginationProvider from "~/components/PaginationProvider";
import { WorkflowVariableSchemaType } from "~/api/tree/schema/workflowVariable";
import { triggerDownloadFromBlob } from "~/util/helpers";
import { useDeleteWorkflowVariable } from "~/api/tree/mutate/deleteVariable";
import { useDownloadVar } from "~/api/tree/mutate/downloadVariable";
import { useTranslation } from "react-i18next";
import { useWorkflowVariables } from "~/api/tree/query/variables";

const pageSize = 10;

const VariablesList = ({ path }: { path: string }) => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteItem, setDeleteItem] = useState<WorkflowVariableSchemaType>();
  const [editItem, setEditItem] = useState<WorkflowVariableSchemaType>();
  const [createItem, setCreateItem] = useState(false);
  const [search, setSearch] = useState("");
  const isSearch = search.length > 0;

  const { data, isFetched } = useWorkflowVariables({ path });

  const { mutate: deleteWorkflowVariable } = useDeleteWorkflowVariable({
    onSuccess: () => {
      setDialogOpen(false);
    },
  });

  const { mutate: downloadVar } = useDownloadVar({
    onSuccess: (response, name) => {
      triggerDownloadFromBlob({
        blob: response.blob,
        filename: name,
      });
    },
  });

  const filteredItems = useMemo(
    () =>
      (data?.variables?.results ?? [])?.filter(
        (item) => !isSearch || item.name.includes(search)
      ),
    [data?.variables?.results, isSearch, search]
  );

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteItem(undefined);
      setCreateItem(false);
      setEditItem(undefined);
    }
  }, [dialogOpen]);

  if (!isFetched) return null;

  const download = (name: string) => {
    downloadVar({
      name,
      path,
    });
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <PaginationProvider items={filteredItems} pageSize={pageSize}>
        {({
          currentItems,
          goToFirstPage,
          goToPage,
          goToNextPage,
          goToPreviousPage,
          currentPage,
          pagesList,
          totalPages,
        }) => (
          <>
            <div className="mb-4 flex flex-col gap-4 sm:flex-row">
              <h3 className="flex grow items-center gap-x-2 pb-2 pt-1 font-bold">
                <Braces className="h-5" />
                {t(
                  "pages.explorer.tree.workflow.settings.variables.list.title"
                )}
              </h3>
              <Input
                className="sm:w-60"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  goToFirstPage();
                }}
                placeholder={t(
                  "pages.settings.variables.list.searchPlaceholder"
                )}
              />
              <CreateItemButton
                data-testid="variable-create"
                onClick={() => setCreateItem(true)}
              >
                {t(
                  "pages.explorer.tree.workflow.settings.variables.list.createBtn"
                )}
              </CreateItemButton>
            </div>
            <Card className="mb-4">
              {currentItems.length ? (
                <Table>
                  <TableBody>
                    {currentItems.map((item, i) => (
                      <ItemRow
                        item={item}
                        key={i}
                        onDelete={setDeleteItem}
                        onEdit={() => setEditItem(item)}
                        onDownload={() => download(item.name)}
                      >
                        {item.name}
                      </ItemRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <NoResult icon={Braces}>
                  {t(
                    isSearch
                      ? "pages.explorer.tree.workflow.settings.variables.list.emptySearch"
                      : "pages.explorer.tree.workflow.settings.variables.list.empty"
                  )}
                </NoResult>
              )}
            </Card>
            {totalPages > 1 && (
              <Pagination>
                <PaginationLink
                  data-testid="pagination-btn-left"
                  icon="left"
                  onClick={() => goToPreviousPage()}
                />
                {pagesList.map((page) => (
                  <PaginationLink
                    active={currentPage === page}
                    key={`${page}`}
                    onClick={() => goToPage(page)}
                  >
                    {page}
                  </PaginationLink>
                ))}
                <PaginationLink
                  data-testid="pagination-btn-right"
                  icon="right"
                  onClick={() => goToNextPage()}
                />
              </Pagination>
            )}
          </>
        )}
      </PaginationProvider>

      {deleteItem && path && (
        <Delete
          name={deleteItem.name}
          onConfirm={() => {
            deleteWorkflowVariable({ variable: deleteItem, path });
            setDeleteItem(undefined);
          }}
        />
      )}

      {createItem && path && (
        <Create
          path={path}
          onSuccess={() => {
            setDialogOpen(false);
          }}
        />
      )}
      {editItem && path && (
        <Edit
          item={editItem}
          path={path}
          onSuccess={() => {
            setDialogOpen(false);
          }}
        />
      )}
    </Dialog>
  );
};

export default VariablesList;
