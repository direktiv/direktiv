import { Card } from "~/design/Card";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import { NoPermissions } from "~/design/Table";
import SyncDetail from "./SyncDetail";
import { useParams } from "@tanstack/react-router";
import { useSyncDetail } from "~/api/syncs/query/get";

const Logs = () => {
  const { sync } = useParams({ from: "/n/$namespace/mirror/logs/$sync" });

  const { isAllowed, noPermissionMessage, isFetched } = useSyncDetail(
    sync || ""
  );

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
      <SyncDetail />
    </>
  );
};

export default Logs;
