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
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { Trans, useTranslation } from "react-i18next";
import { useApiActions, useApiKey } from "~/util/store/apiKey";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";
import { useTheme, useThemeActions } from "~/util/store/theme";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Trash } from "lucide-react";
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
      <Card className="p-4">
        <h1>{t("pages.settings.secrets.list.title")}</h1>
        <Table>
          <TableBody>
            {secrets.data?.secrets.results.map((secret, i) => (
              <TableRow key={i}>
                <TableCell>{secret.name}</TableCell>
                <TableCell>
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

      <Card className="p-4">
        <h1>
          theme <span className="font-bold">{theme}</span>
        </h1>
        <div className="flex space-x-5">
          <Button onClick={() => setTheme("dark")}>darkmode</Button>
          <Button className="" onClick={() => setTheme("light")}>
            lightmode
          </Button>
          <Button onClick={() => setTheme(null)}>reset theme</Button>
        </div>
      </Card>
      <Card className="p-4">
        <h1>
          namespace <span className="font-bold">{selectedNamespace}</span>
        </h1>
        <div className="flex space-x-5">
          <Button variant="destructive" onClick={() => setNamespace(null)}>
            reset namespace
          </Button>
        </div>
      </Card>
      <Card className="p-4">
        <h1>
          api key is <span className="font-bold">{apiKey}</span>
        </h1>
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

export default SettingsPage;
