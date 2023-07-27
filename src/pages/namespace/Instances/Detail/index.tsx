import Diagram from "./Diagram";
import Input from "./Input";
import Logs from "./Logs";
import Output from "./Output";
import { pages } from "~/util/router/pages";

const InstancesDetailPage = () => {
  const { instance } = pages.instances.useParams();

  if (!instance) return null;

  return (
    <div className="grid grow grid-cols-2 gap-5 p-5">
      <Input instanceId={instance} />
      <Output />
      <Logs />
      <Diagram />
    </div>
  );
};

export default InstancesDetailPage;
