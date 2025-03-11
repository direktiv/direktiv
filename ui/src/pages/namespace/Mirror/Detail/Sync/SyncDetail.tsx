import Header from "./Header";
import Logs from "./Logs";
import { useParams } from "@tanstack/react-router";

const SyncDetail = () => {
  const { sync: syncId } = useParams({
    from: "/n/$namespace/mirror/logs/$sync",
  });

  return (
    <div className="flex grow flex-col">
      <Header syncId={syncId} />
      <Logs syncId={syncId} />
    </div>
  );
};

export default SyncDetail;
