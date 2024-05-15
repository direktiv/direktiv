import { FC } from "react";
import VariablesList from "./Variables";
import { usePages } from "~/util/router/pages";

const WorkflowSettingsPage: FC = () => {
  const pages = usePages();
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
