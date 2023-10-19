import { FC } from "react";
import VariablesList from "./Variables";
import { pages } from "~/util/router/pages";

const WorkflowSettingsPage: FC = () => {
  const { path } = pages.explorer.useParams();

  return (
    <div className="flex flex-col space-y-10 p-5">
      <section data-testid="variables-section">
        {path && <VariablesList path={path} />}
      </section>
    </div>
  );
};

export default WorkflowSettingsPage;
