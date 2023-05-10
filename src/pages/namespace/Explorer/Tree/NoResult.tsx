import {
  Dialog,
  DialogContent,
  DialogTrigger,
} from "../../../../design/Dialog";
import { FC, useEffect, useState } from "react";
import { Folder, FolderOpen, Play } from "lucide-react";

import Button from "../../../../design/Button";
import NewDirectory from "./NewDirectory";
import NewWorkflow from "./NewWorkflow";
import { pages } from "../../../../util/router/pages";
import { useNamespace } from "../../../../util/store/namespace";
import { useTranslation } from "react-i18next";

const NoResult: FC = () => {
  const namespace = useNamespace();
  const { path } = pages.explorer.useParams();
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedDialog, setSelectedDialog] = useState<
    "new-dir" | "new-workflow" | undefined
  >();

  useEffect(() => {
    if (dialogOpen === false) setSelectedDialog(undefined);
  }, [dialogOpen, selectedDialog]);

  return (
    <div className="flex flex-col items-center gap-y-5 p-10">
      <div className="flex flex-col items-center justify-center gap-1">
        <FolderOpen />
        <span className="text-center text-sm">
          {t("pages.explorer.tree.list.empty.title")}
        </span>
      </div>
      <div className="flex flex-col gap-5 sm:flex-row">
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger
            asChild
            onClick={() => {
              setSelectedDialog("new-workflow");
            }}
          >
            <Button>
              <Play />
              {t("pages.explorer.tree.list.empty.createWorkflow")}
            </Button>
          </DialogTrigger>
          <DialogTrigger
            asChild
            onClick={() => {
              setSelectedDialog("new-dir");
            }}
          >
            <Button variant="outline">
              <Folder />
              {t("pages.explorer.tree.list.empty.createDirectory")}
            </Button>
          </DialogTrigger>

          <DialogContent>
            {selectedDialog === "new-dir" && (
              <NewDirectory path={path} close={() => setDialogOpen(false)} />
            )}
            {selectedDialog === "new-workflow" && (
              <NewWorkflow path={path} close={() => setDialogOpen(false)} />
            )}
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

export default NoResult;
