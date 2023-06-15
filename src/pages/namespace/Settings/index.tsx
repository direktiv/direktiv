import { FC } from "react";
import RegistriesList from "./Registries";
import SecretsList from "./Secrets";
import VariablesList from "./Variables";

const SettingsPage: FC = () => (
  <div className="flex flex-col space-y-6 p-10">
    <SecretsList />
    <RegistriesList />
    <VariablesList />
  </div>
);

export default SettingsPage;
