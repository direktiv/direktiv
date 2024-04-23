import Header from "./Header";
import Logs from "./Logs";

const SyncDetail = ({ syncId }: { syncId: string }) => (
  <div className="flex grow flex-col">
    <Header syncId={syncId} />
    <Logs syncId={syncId} />
  </div>
);

export default SyncDetail;
