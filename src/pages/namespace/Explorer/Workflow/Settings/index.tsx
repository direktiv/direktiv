import { FC } from "react";
import VariablesList from "./Variables";

const WorkflowSettingsPage: FC = () => (
  <div className="flex flex-col space-y-10 p-5">
    <section data-testid="variables-section">
      <VariablesList />
    </section>
  </div>
);

export default WorkflowSettingsPage;
