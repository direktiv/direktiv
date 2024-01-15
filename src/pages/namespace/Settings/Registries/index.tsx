import { FC, useEffect, useMemo, useState } from "react";
import { NoPermissions, NoResult, Table, TableBody } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Card } from "~/design/Card";
import { Container } from "lucide-react";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import Input from "~/design/Input";
import ItemRow from "../components/ItemRow";
import PaginationProvider from "~/components/PaginationProvider";
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

  const { data, isFetched, isAllowed, noPermissionMessage } = useRegistries();

  const filteredItems = useMemo(
    () =>
      (data?.data ?? [])?.filter(
        (item) => !isSearch || item.url.includes(search)
      ),
    [data?.data, isSearch, search]
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
              {isAllowed ? (
                <>
                  {currentItems.length ? (
                    <Table>
                      <TableBody>
                        {currentItems.map((item) => (
                          <ItemRow
                            key={item.id}
                            item={{
                              ...item,
                              /**
                               * ItemRow is used by registries, secrets and variables
                               * when all migrate to api V2 and have a unique id the
                               * type signature of item can be updated from name to id
                               * (all Playwright tests need to be updated as well then)
                               */
                              name: item.url,
                            }}
                            onDelete={setDeleteRegistry}
                          >
                            {item.url}
                          </ItemRow>
                        ))}
                      </TableBody>
                    </Table>
                  ) : (
                    <NoResult icon={Container}>
                      {t(
                        isSearch
                          ? "pages.settings.registries.list.emptySearch"
                          : "pages.settings.registries.list.empty"
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
      {deleteRegistry && (
        <Delete
          name={deleteRegistry.url}
          onConfirm={() => deleteRegistryMutation({ registry: deleteRegistry })}
        />
      )}
      {createRegistry && <Create onSuccess={() => setDialogOpen(false)} />}
    </Dialog>
  );
};

export default RegistriesList;
