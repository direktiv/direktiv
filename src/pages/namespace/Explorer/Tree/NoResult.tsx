import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { Folder, FolderOpen, Play } from "lucide-react";

import Button from "~/design/Button";
import NewDirectory from "./NewDirectory";
import NewWorkflow from "./NewWorkflow";
import { NoResult as NoResultContainer } from "~/design/Table";
import { pages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const EmptyDirectoryButton = () => {
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
  );
};

const NoResult: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center gap-y-5">
      <NoResultContainer icon={FolderOpen} button={<EmptyDirectoryButton />}>
        {t("pages.explorer.tree.list.empty.title")}
      </NoResultContainer>
    </div>
  );
};

export default NoResult;
