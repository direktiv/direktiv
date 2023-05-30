import List from "./List";
import TrafficShaping from "./TrafficShaping";

const WorkflowRevisionsPage = () => (
  <div className="flex flex-col gap-y-4 p-5 ">
    <TrafficShaping />
    <List />
  </div>
);

export default WorkflowRevisionsPage;
