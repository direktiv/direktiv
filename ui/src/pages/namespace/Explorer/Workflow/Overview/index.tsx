import { FC } from "react";
import Instances from "./Instances";
import Metrics from "./Metrics";
import Services from "./Services";
import { useParams } from "@tanstack/react-router";

const WorkflowOverviewPage: FC = () => {
  const { _splat: path } = useParams({ strict: false });

  if (!path) return null;

  return (
    <div className="grid gap-5 p-4 md:grid-cols-[2fr_1fr]">
      <Instances workflow={path} />
      <Metrics workflow={path} />
      <Services workflow={path} />
    </div>
  );
};

export default WorkflowOverviewPage;
