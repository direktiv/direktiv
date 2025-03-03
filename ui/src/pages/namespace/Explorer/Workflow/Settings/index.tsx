import { FC } from "react";
import VariablesList from "./Variables";
import { useParams } from "@tanstack/react-router";

const WorkflowSettingsPage: FC = () => {
  const { _splat: path } = useParams({ strict: false });

  return (
    <div className="flex flex-col space-y-10 p-5">
      <section data-testid="variables-section">
        {path && <VariablesList path={path} />}
      </section>
    </div>
  );
};

export default WorkflowSettingsPage;
