import { FC } from "react";
import RegistriesList from "./Registries";
import SecretsList from "./Secrets";
import VariablesList from "./Variables";

const SettingsPage: FC = () => (
  <div className="flex flex-col space-y-6 p-10">
    <section data-testid="secrets-section">
      <SecretsList />
    </section>

    <section data-testid="registries-section">
      <RegistriesList />
    </section>

    <section data-testid="variables-section">
      <VariablesList />
    </section>
  </div>
);

export default SettingsPage;
