import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { FC, useState } from "react";
import {
  FolderTree,
  GitCommit,
  Hexagon,
  Key,
  Palette,
  SquareAsterisk,
  Trash,
} from "lucide-react";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { Trans, useTranslation } from "react-i18next";
import { useApiActions, useApiKey } from "~/util/store/apiKey";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";
import { useTheme, useThemeActions } from "~/util/store/theme";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { useDeleteSecret } from "~/api/secrets/mutate/deleteSecret";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useSecrets } from "~/api/secrets/query/get";
import { useVersion } from "~/api/version";

const SettingsPage: FC = () => {
  const apiKey = useApiKey();
  const { setApiKey } = useApiActions();
  const theme = useTheme();
  const { setTheme } = useThemeActions();
  const { setNamespace } = useNamespaceActions();

  const selectedNamespace = useNamespace();

  const secrets = useSecrets();
  const { data: version, isLoading: isVersionLoading } = useVersion();
  const { data: namespaces, isLoading: isLoadingNamespaces } =
    useListNamespaces();

  const [dialogOpen, setDialogOpen] = useState(false);

  const { mutate: deleteSecret } = useDeleteSecret({
    onSuccess: () => {
      setDialogOpen(false);
    },
  });

  const { t } = useTranslation();

  return (
    <div className="flex flex-col space-y-5 p-10">
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <SquareAsterisk className="h-5" />
        {t("pages.settings.secrets.list.title")}
      </h3>

      <Card>
        <Table>
          <TableBody>
            {secrets.data?.secrets.results.map((secret, i) => (
              <TableRow key={i}>
                <TableCell>{secret.name}</TableCell>
                <TableCell className="w-0">
                  <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                    <DialogTrigger data-testid="secret-delete">
                      <Button variant="ghost">
                        <Trash />
                      </Button>
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>
                          <Trash /> Delete
                        </DialogTitle>
                      </DialogHeader>
                      <div className="my-3">
                        <Trans
                          i18nKey="pages.settings.secrets.delete.description"
                          values={{ name: secret.name }}
                        />
                      </div>
                      <DialogFooter>
                        <DialogClose asChild>
                          <Button variant="ghost">Cancel</Button>
                        </DialogClose>
                        <Button
                          data-testid="node-delete-confirm"
                          onClick={() => deleteSecret({ secret })}
                          variant="destructive"
                        >
                          Delete
                        </Button>
                      </DialogFooter>
                    </DialogContent>
                  </Dialog>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
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

      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
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

      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
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
      <div>
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <GitCommit className="h-5" />
          {t("pages.settings.version.title")}
        </h3>
        <div className="mt-2 ml-2">
          {isVersionLoading ? "Loading version...." : version?.api}
        </div>
      </div>
      <div>
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
      </div>
    </div>
  );
};

export default SettingsPage;
