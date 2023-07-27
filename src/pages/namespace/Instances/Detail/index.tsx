import Diagram from "./Diagram";
import Input from "./Input";
import Logs from "./Logs";
import Output from "./Output";

const InstancesDetailPage = () => (
  <div className="grid grow grid-cols-2 gap-5 p-5">
    <Input />
    <Output />
    <Logs />
    <Diagram />
  </div>
);

export default InstancesDetailPage;
