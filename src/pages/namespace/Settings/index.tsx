import { useApiActions, useApiKey } from "../../../util/store/apiKey";
import {
  useNamespace,
  useNamespaceActions,
} from "../../../util/store/namespace";
import { useTheme, useThemeActions } from "../../../util/store/theme";

import { FC } from "react";
import { useListNamespaces } from "../../../api/namespaces/query/get";
import { useVersion } from "../../../api/version";

const SettiongsPage: FC = () => {
  const apiKey = useApiKey();
  const { setApiKey } = useApiActions();
  const theme = useTheme();
  const { setTheme } = useThemeActions();
  const { setNamespace } = useNamespaceActions();

  const selectedNamespace = useNamespace();

  const { data: version, isLoading: isVersionLoading } = useVersion();
  const { data: namespaces, isLoading: isLoadingNamespaces } =
    useListNamespaces();

  return (
    <div className="flex flex-col space-y-5 p-10">
      <div className="card w-max bg-gray-1 p-4 shadow-md dark:bg-gray-dark-1">
        <h1>
          theme <span className="font-bold">{theme}</span>
        </h1>
        <div className="flex space-x-5">
          <button className="btn-primary btn" onClick={() => setTheme("dark")}>
            darkmode
          </button>
          <button className="btn-primary btn" onClick={() => setTheme("light")}>
            lightmode
          </button>
          <button className="btn-error btn" onClick={() => setTheme(null)}>
            reset theme
          </button>
        </div>
      </div>
      <div className="card w-max bg-gray-1 p-4 shadow-md dark:bg-gray-dark-1">
        <h1>
          namespace <span className="font-bold">{selectedNamespace}</span>
        </h1>
        <div className="flex space-x-5">
          <button className="btn-error btn" onClick={() => setNamespace(null)}>
            reset namespace
          </button>
        </div>
      </div>
      <div className="card w-max bg-gray-1 p-4 shadow-md dark:bg-gray-dark-1">
        <h1>
          api key is <span className="font-bold">{apiKey}</span>
        </h1>
        <div className="flex space-x-5">
          <button
            className="btn-primary btn"
            onClick={() => setApiKey("password")}
          >
            set Api key to password
          </button>
          <button className="btn-error btn" onClick={() => setApiKey(null)}>
            reset api key
          </button>
        </div>
      </div>
      <div>
        <h1 className="font-bold">Version</h1>
        {isVersionLoading ? "Loading version...." : version?.api}
      </div>
      <div>
        <h1 className="font-bold">namespaces</h1>
        {isLoadingNamespaces
          ? "Loading namespaces"
          : namespaces?.results.map((namespace) => (
              <div key={namespace.name}>{namespace.name}</div>
            ))}
      </div>

      <div></div>
    </div>
  );
};

export default SettiongsPage;
