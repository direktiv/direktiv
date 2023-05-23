import { FolderTree, GitCommit, Hexagon, Key, Palette } from "lucide-react";
import { useApiActions, useApiKey } from "~/util/store/apiKey";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";
import { useTheme, useThemeActions } from "~/util/store/theme";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import RegistriesList from "./RegistriesList";
import SecretsList from "./SecretsList";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useTranslation } from "react-i18next";
import { useVersion } from "~/api/version";

const SettingsPage: FC = () => {
  const apiKey = useApiKey();
  const { setApiKey } = useApiActions();
  const theme = useTheme();
  const { setTheme } = useThemeActions();
  const { setNamespace } = useNamespaceActions();

  const selectedNamespace = useNamespace();

  const { data: version, isLoading: isVersionLoading } = useVersion();
  const { data: namespaces, isLoading: isLoadingNamespaces } =
    useListNamespaces();

  const { t } = useTranslation();

  return (
    <div className="flex flex-col space-y-6 p-10">
      <section>
        <SecretsList />
      </section>

      <section>
        <RegistriesList />
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Palette className="h-5" />
          {t("pages.settings.theme.title")} {theme}
        </h3>

        <Card className="flex gap-x-3 p-4">
          <div className="flex space-x-5">
            <Button onClick={() => setTheme("dark")}>darkmode</Button>
            <Button className="" onClick={() => setTheme("light")}>
              lightmode
            </Button>
            <Button onClick={() => setTheme(null)}>reset theme</Button>
          </div>
        </Card>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Hexagon className="h-5" />
          {t("pages.settings.namespace.title")} {selectedNamespace}
        </h3>

        <Card className="flex gap-x-3 p-4">
          <div className="flex space-x-5">
            <Button variant="destructive" onClick={() => setNamespace(null)}>
              reset namespace
            </Button>
          </div>
        </Card>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Key className="h-5" />
          {t("pages.settings.apiKey.title")} {apiKey}
        </h3>

        <Card className="flex gap-x-3 p-4">
          <div className="flex space-x-5">
            <Button onClick={() => setApiKey("password")}>
              set Api key to password
            </Button>
            <Button variant="destructive" onClick={() => setApiKey(null)}>
              reset api key
            </Button>
          </div>
        </Card>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <GitCommit className="h-5" />
          {t("pages.settings.version.title")}
        </h3>
        <div className="mt-2 ml-2">
          {isVersionLoading ? "Loading version...." : version?.api}
        </div>
      </section>
      <section>
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <FolderTree className="h-5" />
          {t("pages.settings.namespacesList.title")}
        </h3>
        <div className="mt-2 ml-2">
          {isLoadingNamespaces
            ? "Loading namespaces"
            : namespaces?.results.map((namespace) => (
                <div key={namespace.name}>{namespace.name}</div>
              ))}
        </div>
      </section>
    </div>
  );
};

export default SettingsPage;
