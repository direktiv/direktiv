import { useMemo, useState } from "react";

import { Card } from "~/design/Card";
import Input from "~/design/Input";
import RefreshButton from "~/design/RefreshButton";
import RoutesTable from "./Table";
import { t } from "i18next";
import { useRoutes } from "~/api/gateway/query/getRoutes";

const RoutesPage = () => {
  const { isFetching, refetch } = useRoutes();
  const [search, setSearch] = useState("");
  const { data: routes } = useRoutes();

  const isSearch = search.length > 0;

  const filteredRoutes = useMemo(
    () =>
      (routes?.data ?? [])?.filter(
        (route) =>
          !isSearch ||
          route.file_path.includes(search) ||
          route.server_path?.includes(search)
      ),
    [isSearch, search, routes?.data]
  );

  return (
    <Card className="m-5">
      <div className="flex justify-between gap-5 p-2">
        <Input
          className="sm:w-60"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
          }}
          placeholder={t("pages.gateway.routes.searchPlaceholder")}
        />
        <RefreshButton
          icon
          variant="outline"
          disabled={isFetching}
          onClick={() => {
            refetch();
          }}
        />
      </div>
      <RoutesTable search={search} filteredRoutes={filteredRoutes} />
    </Card>
  );
};

export default RoutesPage;
