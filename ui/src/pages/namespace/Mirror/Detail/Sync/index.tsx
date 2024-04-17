import { Card } from "~/design/Card";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { NoPermissions } from "~/design/Table";
import SyncDetail from "./SyncDetail";
import { pages } from "~/util/router/pages";
import { useSyncDetail } from "~/api/syncs/query/get";

const Logs = () => {
  const { sync } = pages.mirror.useParams();
  const { isAllowed, noPermissionMessage, isFetched } = useSyncDetail(
    sync || ""
  );

  if (!sync) return null;
  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  return (
    <>
      <LogStreamingSubscriber activity={sync} />
      <SyncDetail syncId={sync} />
    </>
  );
};

export default Logs;
