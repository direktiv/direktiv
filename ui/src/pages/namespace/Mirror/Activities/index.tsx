import { FolderSync, GitCompare } from "lucide-react";
import {
  NoPermissions,
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Card } from "~/design/Card";
import Header from "./Header";
import PaginationProvider from "~/components/PaginationProvider";
import Row from "./Row";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceDetail } from "~/api/namespaces/query/get";
import { useQueryClient } from "@tanstack/react-query";
import { useSyncs } from "~/api/syncs/query/get";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const Activities = () => {
  const { data, isAllowed, noPermissionMessage, isFetched } = useSyncs();
  const namespace = useNamespace();
  const namespaceDetail = useNamespaceDetail();

  const mirror = namespaceDetail.data?.mirror;

  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const apiKey = useApiKey();

  const syncs = data?.data;
  const noResults = isFetched && data?.data.length === 0;

  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  if (!namespace) return null;
  if (!mirror) return null;
  if (!syncs) return null;

  // const refreshActivities = () => {
  //   queryClient.invalidateQueries(
  //     treeKeys.mirrorInfo(data.namespace, {
  //       apiKey: apiKey ?? undefined,
  //     })
  //   );
  // };

  const pendingActivities = syncs.filter((sync) => sync.status === "executing");

  // if (pendingActivities.length) {
  //   setTimeout(() => refreshActivities(), 1000);
  // }

  return (
    <>
      <Header mirror={mirror} loading={!!pendingActivities.length} />
      <PaginationProvider items={syncs} pageSize={pageSize}>
        {({
          currentItems,
          goToPage,
          goToNextPage,
          goToPreviousPage,
          currentPage,
          pagesList,
          totalPages,
        }) => (
          <div className="flex grow flex-col gap-y-4 p-5">
            <h3 className="flex items-center gap-x-2 font-bold">
              <FolderSync className="h-5" />
              {t("pages.mirror.syncs.list.title")}
            </h3>
            <Card>
              <Table className="border-gray-5 dark:border-gray-dark-5">
                <TableHead>
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableHeaderCell>
                      {t("pages.mirror.syncs.tableHeader.id")}
                    </TableHeaderCell>
                    <TableHeaderCell>
                      {t("pages.mirror.syncs.tableHeader.type")}
                    </TableHeaderCell>
                    <TableHeaderCell>
                      {t("pages.mirror.syncs.tableHeader.status")}
                    </TableHeaderCell>
                    <TableHeaderCell>
                      {t("pages.mirror.syncs.tableHeader.createdAt")}
                    </TableHeaderCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {noResults && (
                    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                      <TableCell colSpan={4}>
                        <NoResult icon={GitCompare}>
                          {t("pages.mirror.syncs.list.noResults")}
                        </NoResult>
                      </TableCell>
                    </TableRow>
                  )}
                  {currentItems.map((sync) => (
                    <Row namespace={namespace} key={sync.id} item={sync} />
                  ))}
                </TableBody>
              </Table>
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
          </div>
        )}
      </PaginationProvider>
    </>
  );
};

export default Activities;
