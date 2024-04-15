import Header from "./Header";
import Logs from "./Logs";

const ActivityDetail = ({ syncId }: { syncId: string }) => (
  <div className="flex grow flex-col">
    <Header syncId={syncId} />
    <Logs syncId={syncId} />
  </div>
);

export default ActivityDetail;
