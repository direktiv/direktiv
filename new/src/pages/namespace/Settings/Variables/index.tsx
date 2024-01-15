import { FC, useEffect, useMemo, useState } from "react";
import { NoPermissions, NoResult, Table, TableBody } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Braces } from "lucide-react";
import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import Edit from "./Edit";
import Input from "~/design/Input";
import ItemRow from "../components/ItemRow";
import PaginationProvider from "~/componentsNext/PaginationProvider";
import { VarSchemaType } from "~/api/variables/schema";
import { triggerDownloadFromBlob } from "~/util/helpers";
import { useDeleteVar } from "~/api/variables/mutate/deleteVariable";
import { useDownloadVar } from "~/api/variables/mutate/downloadVariable";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/variables/query/useVariables";

const pageSize = 10;

const VariablesList: FC = () => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteItem, setDeleteItem] = useState<VarSchemaType>();
  const [editItem, setEditItem] = useState<VarSchemaType>();
  const [createItem, setCreateItem] = useState(false);
  const [search, setSearch] = useState("");
  const isSearch = search.length > 0;

  const { data, isFetched, isAllowed, noPermissionMessage } = useVars();

  const filteredItems = useMemo(
    () =>
      (data?.variables?.results ?? [])?.filter(
        (item) => !isSearch || item.name.includes(search)
      ),
    [data?.variables?.results, isSearch, search]
  );

  const { mutate: deleteVarMutation } = useDeleteVar({
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

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteItem(undefined);
      setCreateItem(false);
      setEditItem(undefined);
    }
  }, [dialogOpen]);

  if (!isFetched) return null;

  const download = (name: string) => {
    downloadVar(name);
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
          name={deleteItem.name}
          onConfirm={() => {
            deleteVarMutation({ variable: deleteItem });
          }}
        />
      )}
      {createItem && (
        <Create
          onSuccess={() => {
            setDialogOpen(false);
          }}
        />
      )}
      {editItem && (
        <Edit
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
