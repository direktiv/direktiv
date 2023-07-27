import Diagram from "./Diagram";
import Header from "./Header";
import Input from "./Input";
import Logs from "./Logs";
import Output from "./Output";
import { pages } from "~/util/router/pages";

const InstancesDetailPage = () => {
  const { instance } = pages.instances.useParams();

  if (!instance) return null;

  return (
    <div className="flex grow flex-col">
      <Header instanceId={instance} />
      <div className="grid grow grid-cols-2 gap-5 p-5">
        <Logs />
        <Diagram />
        <Input instanceId={instance} />
        <Output instanceId={instance} />
      </div>
    </div>
  );
};

export default InstancesDetailPage;
