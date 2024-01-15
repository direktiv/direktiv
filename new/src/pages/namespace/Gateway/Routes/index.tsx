import { Card } from "~/design/Card";
import RefreshButton from "~/design/RefreshButton";
import RoutesTable from "./Table";
import { useRoutes } from "~/api/gateway/query/getRoutes";

const RoutesPage = () => {
  const { isFetching, refetch } = useRoutes();

  return (
    <Card className="m-5">
      <div className="flex justify-end gap-5 p-2">
        <RefreshButton
          icon
          variant="outline"
          disabled={isFetching}
          onClick={() => {
            refetch();
          }}
        />
      </div>
      <RoutesTable />
    </Card>
  );
};

export default RoutesPage;
