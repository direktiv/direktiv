import { FC, useEffect, useState } from "react";
import { Pagination, PaginationLink } from "~/design/Pagination";
import { Table, TableBody } from "~/design/Table";

import { Braces } from "lucide-react";
import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import Edit from "./Edit";
import EmptyList from "../components/EmptyList";
import Input from "~/design/Input";
import ItemRow from "../components/ItemRow";
import PaginationProvider from "~/componentsNext/PaginationProvider";
import { VarSchemaType } from "~/api/variables/schema";
import { useDeleteVar } from "~/api/variables/mutate/deleteVariable";
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

  const { data, isFetched } = useVars();
  const items = data?.variables?.results ?? null;

  // wrap in useMemo
  const filteredItems = items?.filter(
    (item) => !search || item.name.includes(search)
  );

  const { mutate: deleteVarMutation } = useDeleteVar({
    onSuccess: () => {
      setDialogOpen(false);
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

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <PaginationProvider items={filteredItems ?? []} pageSize={pageSize}>
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
            <div className="mb-3 flex flex-row gap-3">
              <h3 className="flex grow items-center gap-x-2 pb-2 pt-1 font-bold">
                <Braces className="h-5" />
                {t("pages.settings.variables.list.title")}
              </h3>

              <Input
                className="w-1/3"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  goToFirstPage();
                }}
                placeholder="search variables:"
              />
              <CreateItemButton
                data-testid="variable-create"
                onClick={() => setCreateItem(true)}
              />
            </div>

            <Card className="mb-3">
              {currentItems?.length ? (
                <Table>
                  <TableBody>
                    {currentItems?.map((item, i) => (
                      <ItemRow
                        item={item}
                        key={i}
                        onDelete={setDeleteItem}
                        onEdit={() => setEditItem(item)}
                      />
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <EmptyList>
                  {t("pages.settings.variables.list.empty")}
                </EmptyList>
              )}
            </Card>
            {totalPages > 2 && (
              <Pagination>
                <PaginationLink
                  icon="left"
                  onClick={() => goToPreviousPage()}
                />
                {pagesList.map((p) => (
                  <PaginationLink
                    active={currentPage === p}
                    key={`${p}`}
                    onClick={() => goToPage(p)}
                  >
                    {p}
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
          onConfirm={() => deleteVarMutation({ variable: deleteItem })}
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
