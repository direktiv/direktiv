import { FC, useEffect, useMemo, useState } from "react";
import { Pagination, PaginationLink } from "~/design/Pagination";
import { Table, TableBody } from "~/design/Table";

import { Card } from "~/design/Card";
import Create from "./Create";
import CreateItemButton from "../components/CreateItemButton";
import Delete from "./Delete";
import { Dialog } from "~/design/Dialog";
import EmptyList from "../components/EmptyList";
import Input from "~/design/Input";
import ItemRow from "../components/ItemRow";
import PaginationProvider from "~/componentsNext/PaginationProvider";
import { SecretSchemaType } from "~/api/secrets/schema";
import { SquareAsterisk } from "lucide-react";
import { useDeleteSecret } from "~/api/secrets/mutate/deleteSecret";
import { useSecrets } from "~/api/secrets/query/get";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const SecretsList: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteSecret, setDeleteSecret] = useState<SecretSchemaType>();
  const [createSecret, setCreateSecret] = useState(false);
  const [search, setSearch] = useState("");
  const isSearch = search.length > 0;

  const { data, isFetched } = useSecrets();

  const filteredItems = useMemo(
    () =>
      (data?.secrets.results ?? [])?.filter(
        (item) => !isSearch || item.name.includes(search)
      ),
    [data?.secrets.results, isSearch, search]
  );

  const { mutate: deleteSecretMutation } = useDeleteSecret({
    onSuccess: () => {
      setDeleteSecret(undefined);
      setDialogOpen(false);
    },
  });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteSecret(undefined);
      setCreateSecret(false);
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
                <SquareAsterisk className="h-5" />
                {t("pages.settings.secrets.list.title")}
              </h3>
              <Input
                className="sm:w-60"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  goToFirstPage();
                }}
                placeholder={t("pages.settings.secrets.list.searchPlaceholder")}
              />
              <CreateItemButton
                data-testid="secret-create"
                onClick={() => setCreateSecret(true)}
              >
                {t("pages.settings.secrets.list.createBtn")}
              </CreateItemButton>
            </div>
            <Card className="mb-4">
              {currentItems.length ? (
                <Table>
                  <TableBody>
                    {currentItems.map((item, i) => (
                      <ItemRow item={item} key={i} onDelete={setDeleteSecret} />
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <EmptyList>
                  {t(
                    isSearch
                      ? "pages.settings.secrets.list.emptySearch"
                      : "pages.settings.secrets.list.empty"
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
      {deleteSecret && (
        <Delete
          name={deleteSecret.name}
          onConfirm={() => deleteSecretMutation({ secret: deleteSecret })}
        />
      )}
      {createSecret && <Create onSuccess={() => setDialogOpen(false)} />}
    </Dialog>
  );
};

export default SecretsList;
