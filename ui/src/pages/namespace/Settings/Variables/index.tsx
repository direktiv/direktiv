import { FC, useEffect, useMemo, useState } from "react";
import { FileJson, Trash } from "lucide-react";
import { NoPermissions, NoResult, Table, TableBody } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Checkbox } from "~/design/Checkbox";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import Edit from "./Edit";
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
  const [editItem, setEditItem] = useState<VarSchemaType>();
  const [createItem, setCreateItem] = useState(false);
  const [selectedItems, setSelectedItems] = useState<VarSchemaType[]>([]);
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
      setSelectedItems([]);
      setCreateItem(false);
      setEditItem(undefined);
    }
  }, [dialogOpen]);

  if (!isFetched) return null;

  const download = (variableId: string) => {
    downloadVar(variableId);
  };

  const handleCheckboxChange = (item: VarSchemaType) => {
    setSelectedItems((prevSelected) =>
      prevSelected.some((selected) => selected.id === item.id)
        ? prevSelected.filter((selected) => selected.id !== item.id)
        : [...prevSelected, item]
    );
  };

  const handleSelectAll = () => {
    if (selectedItems.length === filteredItems.length) {
      setSelectedItems([]);
    } else {
      setSelectedItems(filteredItems);
    }
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
            </div>
            <Card className="mb-4">
              <div className="flex justify-between gap-5 p-2 border-b border-gray-5 dark:border-gray-dark-5">
                <div className="flex items-center">
                  {currentItems.length > 0 && (
                    <Checkbox
                      className="ml-1"
                      onCheckedChange={handleSelectAll}
                      checked={
                        selectedItems.length === filteredItems.length &&
                        filteredItems.length > 0
                      }
                    />
                  )}
                </div>
                <div className="mr-auto">
                  {selectedItems.length > 0 && !dialogOpen && (
                    <Button
                      variant="destructive"
                      disabled={selectedItems.length === 0}
                      onClick={() => {
                        setDialogOpen(true);
                      }}
                    >
                      <Trash className=" size-4" />
                      {t(
                        "pages.explorer.tree.workflow.settings.variables.list.deleteSelected"
                      )}
                    </Button>
                  )}
                </div>
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

              {isAllowed ? (
                <>
                  {currentItems.length ? (
                    <Table>
                      <TableBody>
                        {currentItems.map((item, i) => (
                          <ItemRow
                            item={item}
                            key={i}
                            onDelete={() => handleCheckboxChange(item)}
                            onEdit={() => setEditItem(item)}
                            onDownload={() => download(item.id)}
                            onSelect={() => handleCheckboxChange(item)}
                            isSelected={selectedItems.some(
                              (selected) => selected.id === item.id
                            )}
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
            <Pagination>
              <PaginationLink
                disabled={totalPages === 1}
                icon="left"
                onClick={() => goToPreviousPage()}
              />
              {pagesList.map((page) => (
                <PaginationLink
                  disabled={totalPages === 1}
                  active={currentPage === page}
                  key={`${page}`}
                  onClick={() => goToPage(page)}
                >
                  {page}
                </PaginationLink>
              ))}
              <PaginationLink
                disabled={totalPages === 1}
                icon="right"
                onClick={() => goToNextPage()}
              />
            </Pagination>
          </>
        )}
      </PaginationProvider>

      {selectedItems.length > 0 && (
        <Delete
          items={selectedItems}
          totalItems={variables?.data?.length ?? 0}
          onConfirm={() => {
            deleteVar(selectedItems);
            setSelectedItems([]);
            setDialogOpen(false);
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
