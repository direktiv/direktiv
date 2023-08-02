import ApiKey from "./ApiKey";
import Broadcasts from "./Broadcasts";
import { FC } from "react";
import Input from "~/design/Input";
import RegistriesList from "./Registries";
import SecretsList from "./Secrets";
import VariablesList from "./Variables";
import { useBroadcasts } from "~/api/broadcasts/query/useBroadcasts";
import { useRegistries } from "~/api/registries/query/get";
import { useSecrets } from "~/api/secrets/query/get";
import { useVars } from "~/api/variables/query/useVariables";

// this hook uses all initial data fetching hooks from the components on
// this page and returns true if all of them are fetched. This is used to
// render the page only when all data is fetched and avoid flickering as
// all components are rendered at the same time and data comes in at different
// times.
const usePrefetchData = () => {
  const { isFetched: isFetchedBroadcasts } = useBroadcasts();
  const { isFetched: isFetchedSecrets } = useSecrets();
  const { isFetched: isFetchedRegistries } = useRegistries();
  const { isFetched: isFetchedVars } = useVars();
  return (
    isFetchedBroadcasts &&
    isFetchedSecrets &&
    isFetchedRegistries &&
    isFetchedVars
  );
};

const SettingsPage: FC = () => {
  const prefetched = usePrefetchData();

  if (!prefetched) {
    return null;
  }

  return (
    <div className="flex flex-col space-y-10 p-5">
      <section data-testid="secrets-section">
        <SecretsList />
      </section>

      <section data-testid="registries-section">
        <RegistriesList />
      </section>

      <section data-testid="variables-section">
        <VariablesList />
      </section>

      <section data-testid="broadcasts-section">
        <Broadcasts />
      </section>

      <section>
        <ApiKey />
      </section>
    </div>
  );
};

export default SettingsPage;
