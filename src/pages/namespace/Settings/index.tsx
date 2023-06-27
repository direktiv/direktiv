import Broadcasts from "./Broadcasts";
import { FC } from "react";
import RegistriesList from "./Registries";
import SecretsList from "./Secrets";
import VariablesList from "./Variables";

const SettingsPage: FC = () => (
  <div className="flex flex-col space-y-6 p-10">
    <section>
      <SecretsList />
    </section>

    <section>
      <RegistriesList />
    </section>

    <section>
      <VariablesList />
    </section>

    <section data-testid="broadcasts-section">
      <Broadcasts />
    </section>
  </div>
);

export default SettingsPage;
