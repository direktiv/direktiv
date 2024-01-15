import List from "./List";
import TrafficShaping from "./TrafficShaping";

const RevisionsOverviewPage = () => (
  <div className="flex flex-col space-y-10">
    <TrafficShaping />
    <List />
  </div>
);

export default RevisionsOverviewPage;
