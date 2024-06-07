import { Book, Github, LogOut, PlusCircle, Slack } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { useEffect, useState } from "react";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import Logo from "~/components/Logo";
import LogoutButton from "~/components/LogoutButton";
import NamespaceEdit from "~/components/NamespaceEdit";
import useApiKeyHandling from "~/hooks/useApiKeyHandling";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNavigate } from "react-router-dom";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const Layout = () => {
  const pages = usePages();
  const { t } = useTranslation();
  const { usesAccounts } = useApiKeyHandling();
  const {
    data: availableNamespaces,
    isFetched,
    isError,
    isRefetching,
  } = useListNamespaces();
  const activeNamespace = useNamespace();
  const { setNamespace } = useNamespaceActions();
  const [, setDialogOpen] = useState(false);

  const navigate = useNavigate();

  const linkItems = [
    {
      icon: <Book />,
      title: t("pages.onboarding.links.docs.title"),
      description: t("pages.onboarding.links.docs.description"),
      href: "https://docs.direktiv.io/getting_started/",
    },
    {
      icon: <Slack />,
      title: t("pages.onboarding.links.slack.title"),
      description: t("pages.onboarding.links.slack.description"),
      href: "https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g",
    },
    {
      icon: <Github />,
      title: t("pages.onboarding.links.github.title"),
      description: t("pages.onboarding.links.github.description"),
      href: "https://github.com/direktiv/direktiv",
    },
  ];

  useEffect(() => {
    if (
      availableNamespaces &&
      availableNamespaces.data[0] &&
      /**
       * the namespace list might still be refetching after a cache invalidation. This could be caused by a
       * namespace delete action that was just triggered. We have to wait until the refetch is done to avoid
       * using an old namespaces list.
       */
      !isRefetching
    ) {
      // if there is a prefered namespace in localStorage, redirect to it
      if (
        activeNamespace &&
        availableNamespaces.data.some((ns) => ns.name === activeNamespace)
      ) {
        navigate(pages.explorer.createHref({ namespace: activeNamespace }));
        return;
      }
      // otherwise, redirect to the first namespace and store it in localStorage
      setNamespace(availableNamespaces.data[0].name);
      navigate(
        pages.explorer.createHref({
          namespace: availableNamespaces.data[0].name,
        })
      );
      return;
    }
  }, [
    activeNamespace,
    availableNamespaces,
    isRefetching,
    navigate,
    pages.explorer,
    setNamespace,
  ]);

  // wait until namespaces are fetched to avoid layout shifts
  // either the useEffect will redirect or the onboarding screen
  // will be shown
  if (!isFetched) {
    return null;
  }

  return (
    <main className="grid min-h-full place-items-center py-24 px-6 sm:py-32 lg:px-8">
      <div className="flex flex-col gap-8 text-center">
        <h1 className="flex justify-center space-x-3 text-2xl font-bold text-gray-12 dark:text-gray-dark-12">
          <span>{t("pages.onboarding.welcomeTo")}</span>
          <Logo />
        </h1>
        {isError && (
          <Alert variant="error">{t("pages.onboarding.error")}</Alert>
        )}
        <div className="relative block w-full rounded-lg border-2 border-dashed border-gray-5 p-12 text-center dark:border-gray-dark-5">
          <p className="mt-1 text-sm text-gray-9 dark:text-gray-dark-9">
            {t("pages.onboarding.noNamespace")}
          </p>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="lg" className="my-5">
                <PlusCircle />
                {t("pages.onboarding.createNamespaceBtn")}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <NamespaceEdit close={() => setDialogOpen(false)} />
            </DialogContent>
          </Dialog>
        </div>
        {usesAccounts && (
          <LogoutButton
            button={(props) => <Button {...props} variant="outline" />}
          >
            <LogOut />
            {t("pages.onboarding.logout")}
          </LogoutButton>
        )}
        <ul role="list" className="text-left">
          {linkItems.map((item, itemIdx) => (
            <li key={itemIdx}>
              <div className="group relative flex items-start space-x-3 py-4">
                <div className="shrink-0">
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg">
                    {item.icon}
                  </span>
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium text-gray-11 dark:text-gray-dark-11">
                    <a
                      href={item.href}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <span className="absolute inset-0" aria-hidden="true" />
                      {item.title}
                    </a>
                  </div>
                  <p className="text-sm text-gray-9 dark:text-gray-dark-9">
                    {item.description}
                  </p>
                </div>
                <div className="shrink-0 self-center"></div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
};

export default Layout;
