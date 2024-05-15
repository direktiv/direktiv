import { FC } from "react";
import Instances from "./Instances";
import Metrics from "./Metrics";
import Services from "./Services";
import { usePages } from "~/util/router/pages";

const ActiveWorkflowPage: FC = () => {
  const pages = usePages();
  const { path } = pages.explorer.useParams();

  if (!path) return null;

  return (
    <div className="grid gap-5 p-4 md:grid-cols-[2fr_1fr]">
      <Instances workflow={path} />
      <Metrics workflow={path} />
      <Services workflow={path} />
    </div>
  );
};

export default ActiveWorkflowPage;
