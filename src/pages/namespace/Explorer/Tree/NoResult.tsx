import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import NewFileButton, { FileTypeSelection } from "./components/NewFileButton";

import { FolderOpen } from "lucide-react";
import NewConsumer from "./CreateNew/Consumer";
import NewDirectory from "./NewDirectory";
import NewRoute from "./NewRoute";
import NewService from "./NewService";
import NewWorkflow from "./NewWorkflow";
import { NoResult as NoResultContainer } from "~/design/Table";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

const EmptyDirectoryButton = () => {
  const { path } = pages.explorer.useParams();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedDialog, setSelectedDialog] = useState<FileTypeSelection>();

  useEffect(() => {
    if (dialogOpen === false) setSelectedDialog(undefined);
  }, [dialogOpen, selectedDialog]);

  const wideOverlay =
    !!selectedDialog &&
    !["new-dir", "new-route", "new-consumer"].includes(selectedDialog);

  return (
    <div className="grid gap-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <NewFileButton setSelectedDialog={setSelectedDialog} />
        <DialogContent
          className={twMergeClsx(
            wideOverlay && "sm:max-w-xl md:max-w-2xl lg:max-w-3xl"
          )}
        >
          {selectedDialog === "new-dir" && (
            <NewDirectory path={path} close={() => setDialogOpen(false)} />
          )}
          {selectedDialog === "new-workflow" && (
            <NewWorkflow path={path} close={() => setDialogOpen(false)} />
          )}
          {selectedDialog === "new-service" && (
            <NewService path={path} close={() => setDialogOpen(false)} />
          )}
          {selectedDialog === "new-route" && (
            <NewRoute path={path} close={() => setDialogOpen(false)} />
          )}
          {selectedDialog === "new-consumer" && (
            <NewConsumer path={path} close={() => setDialogOpen(false)} />
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
