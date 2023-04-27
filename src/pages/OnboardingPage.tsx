import { useNamespace, useNamespaceActions } from "../util/store/namespace";

import Logo from "../design/Logo";
import { pages } from "../util/router/pages";
import { useEffect } from "react";
import { useListNamespaces } from "../api/namespaces/query/get";
import { useNavigate } from "react-router-dom";

const Layout = () => {
  const { data: availableNamespaces, isFetched } = useListNamespaces();
  const activeNamespace = useNamespace();
  const { setNamespace } = useNamespaceActions();

  const navigate = useNavigate();

  useEffect(() => {
    if (availableNamespaces && availableNamespaces.results[0]) {
      // if there is a prefered namespace in localStorage, redirect to it
      if (
        activeNamespace &&
        availableNamespaces.results.some((ns) => ns.name === activeNamespace)
      ) {
        navigate(pages.explorer.createHref({ namespace: activeNamespace }));
        return;
      }
      // otherwise, redirect to the first namespace and store it in localStorage
      setNamespace(availableNamespaces.results[0].name);
      navigate(
        pages.explorer.createHref({
          namespace: availableNamespaces.results[0].name,
        })
      );
      return;
    }
  }, [activeNamespace, availableNamespaces, navigate, setNamespace]);

  // wait until namespaces are fetched to avoid layout shifts
  // either the useEffect will redirect or the onboarding screen
  // will be shown
  if (!isFetched) {
    return null;
  }

  return (
    <main className="grid min-h-full place-items-center bg-white py-24 px-6 sm:py-32 lg:px-8">
      <div className="text-center">
        <h1 className="flex items-center space-x-3 text-2xl font-bold text-gray-900">
          <span>Welcome to </span>
          <Logo />
        </h1>
      </div>
    </main>
  );
};

export default Layout;
