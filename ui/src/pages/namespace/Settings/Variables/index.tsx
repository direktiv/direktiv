import { FC, useEffect, useMemo, useState } from "react";
import { NoPermissions, NoResult, Table, TableBody } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import Edit from "./Edit";
import { FileJson } from "lucide-react";
import Input from "~/design/Input";
import ItemRow from "../components/ItemRow";
import PaginationProvider from "~/components/PaginationProvider";
import { VarSchemaType } from "~/api/variables/schema";
import { triggerDownloadFromBase64String } from "~/util/helpers";
import { useDeleteVar } from "~/api/variables/mutate/delete";
import { useDownloadVar } from "~/api/variables/mutate/download";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/variables/query/get";

const pageSize = 10;

const VariablesList: FC = () => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteItem, setDeleteItem] = useState<VarSchemaType>();
  const [editItem, setEditItem] = useState<VarSchemaType>();
  const [createItem, setCreateItem] = useState(false);
  const [search, setSearch] = useState("");
  const isSearch = search.length > 0;

  const {
    data: variables,
    isFetched,
    isAllowed,
    noPermissionMessage,
  } = useVars();

  const filteredItems = useMemo(
    () =>
      (variables?.data ?? [])?.filter(
        (item) => !isSearch || item.name.includes(search)
      ),
    [isSearch, search, variables?.data]
  );

  const { mutate: deleteVar } = useDeleteVar({
    onSuccess: () => {
      setDialogOpen(false);
    },
  });

  const allNames = variables?.data?.map((v) => v.name) ?? [];

  const { mutate: downloadVar } = useDownloadVar({
    onSuccess: (response) => {
      const { name: filename, data: base64String, mimeType } = response.data;
      triggerDownloadFromBase64String({
        filename,
        base64String,
        mimeType,
      });
    },
  });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteItem(undefined);
      setCreateItem(false);
      setEditItem(undefined);
    }
  }, [dialogOpen]);

  if (!isFetched) return null;

  const download = (variableId: string) => {
    downloadVar(variableId);
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
                <FileJson className="h-5" />
                {t("pages.settings.variables.list.title")}
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
                {t("pages.settings.variables.list.createBtn")}
              </CreateItemButton>
            </div>
            <Card className="mb-4">
              {isAllowed ? (
                <>
                  {currentItems.length ? (
                    <Table>
                      <TableBody>
                        {currentItems.map((item, i) => (
                          <ItemRow
                            item={item}
                            key={i}
                            onDelete={setDeleteItem}
                            onEdit={() => setEditItem(item)}
                            onDownload={() => download(item.id)}
                          >
                            {item.name}
                          </ItemRow>
                        ))}
                      </TableBody>
                    </Table>
                  ) : (
                    <NoResult icon={FileJson}>
                      {t(
                        isSearch
                          ? "pages.settings.variables.list.emptySearch"
                          : "pages.settings.variables.list.empty"
                      )}
                    </NoResult>
                  )}
                </>
              ) : (
                <NoPermissions>{noPermissionMessage}</NoPermissions>
              )}
            </Card>
            {totalPages > 1 && (
              <Pagination>
                <PaginationLink
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
                <PaginationLink icon="right" onClick={() => goToNextPage()} />
              </Pagination>
            )}
          </>
        )}
      </PaginationProvider>

      {deleteItem && (
        <Delete
          items={[deleteItem]}
          totalItems={variables?.data?.length ?? 0}
          onConfirm={() => {
            deleteVar({ variable: deleteItem });
          }}
        />
      )}
      {createItem && (
        <Create
          unallowedNames={allNames}
          onSuccess={() => {
            setDialogOpen(false);
          }}
        />
      )}
      {editItem && (
        <Edit
          unallowedNames={allNames.filter((name) => name !== editItem.name)}
          item={editItem}
          onSuccess={() => {
            setDialogOpen(false);
          }}
        />
      )}
    </Dialog>
  );
};

export default VariablesList;
