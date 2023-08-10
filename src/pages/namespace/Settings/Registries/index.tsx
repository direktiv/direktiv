import { FC, useEffect, useMemo, useState } from "react";
import { Pagination, PaginationLink } from "~/design/Pagination";
import { Table, TableBody } from "~/design/Table";

import { Card } from "~/design/Card";
import { Container } from "lucide-react";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import EmptyList from "../components/EmptyList";
import Input from "~/design/Input";
import ItemRow from "../components/ItemRow";
import PaginationProvider from "~/componentsNext/PaginationProvider";
import { RegistrySchemaType } from "~/api/registries/schema";
import { useDeleteRegistry } from "~/api/registries/mutate/deleteRegistry";
import { useRegistries } from "~/api/registries/query/get";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const RegistriesList: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteRegistry, setDeleteRegistry] = useState<RegistrySchemaType>();
  const [createRegistry, setCreateRegistry] = useState(false);
  const [search, setSearch] = useState("");
  const isSearch = search.length > 0;

  const { data, isFetched } = useRegistries();

  const filteredItems = useMemo(
    () =>
      (data?.registries ?? [])?.filter(
        (item) => !isSearch || item.name.includes(search)
      ),
    [data?.registries, isSearch, search]
  );

  const { mutate: deleteRegistryMutation } = useDeleteRegistry({
    onSuccess: () => {
      setDeleteRegistry(undefined);
      setDialogOpen(false);
    },
  });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteRegistry(undefined);
      setCreateRegistry(false);
    }
  }, [dialogOpen]);

  if (!isFetched) return null;

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
                <Container className="h-5" />
                {t("pages.settings.registries.list.title")}
              </h3>
              <Input
                className="sm:w-60"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  goToFirstPage();
                }}
                placeholder={t(
                  "pages.settings.registries.list.searchPlaceholder"
                )}
              />
              <CreateItemButton
                onClick={() => setCreateRegistry(true)}
                data-testid="registry-create"
              >
                {t("pages.settings.registries.list.createBtn")}
              </CreateItemButton>
            </div>
            <Card className="mb-4">
              {currentItems.length ? (
                <Table>
                  <TableBody>
                    {currentItems.map((item, i) => (
                      <ItemRow
                        key={i}
                        item={item}
                        onDelete={setDeleteRegistry}
                      />
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <EmptyList>
                  {t(
                    isSearch
                      ? "pages.settings.registries.list.emptySearch"
                      : "pages.settings.registries.list.empty"
                  )}
                </EmptyList>
              )}
            </Card>
            {totalPages > 1 && (
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
      {deleteRegistry && (
        <Delete
          name={deleteRegistry.name}
          onConfirm={() => deleteRegistryMutation({ registry: deleteRegistry })}
        />
      )}
      {createRegistry && <Create onSuccess={() => setDialogOpen(false)} />}
    </Dialog>
  );
};

export default RegistriesList;
