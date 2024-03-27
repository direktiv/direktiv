import { Card } from "~/design/Card";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { NoPermissions } from "~/design/Table";
import { pages } from "~/util/router/pages";
import { useRoute } from "~/api/gateway/query/getRoutes";

const RoutesDetailPage = () => {
  const { routePath } = pages.gateway.useParams();
  const { data, isAllowed, isFetched, noPermissionMessage } = useRoute({
    routePath: routePath ?? "",
    enabled: !!routePath,
  });

  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  if (!data) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      {data.path && <LogStreamingSubscriber route={data.path} />}
      Details for {routePath} {data.path}
    </div>
  );
};

export default RoutesDetailPage;
