import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import NewFileButton, { FileTypeSelection } from "./components/NewFileButton";

import NewConsumer from "./components/modals/CreateNew/Gateway/Consumer";
import NewDirectory from "./components/modals/CreateNew/Directory";
import NewRoute from "./components/modals/CreateNew/Gateway/Route";
import NewService from "./components/modals/CreateNew/Service";
import NewWorkflow from "./components/modals/CreateNew/Workflow";
import { getFilenameFromPath } from "~/api/files/schema";
import { twMergeClsx } from "~/util/helpers";
import { useNode } from "~/api/files/query/node";

type NewFileDialogProps = {
  path: string | undefined;
};

export const NewFileDialog: FC<NewFileDialogProps> = ({ path }) => {
  const { data } = useNode({ path });

  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedDialog, setSelectedDialog] = useState<FileTypeSelection>();

  useEffect(() => {
    if (dialogOpen === false) setSelectedDialog(undefined);
  }, [dialogOpen, selectedDialog]);

  const wideOverlay = !!selectedDialog && selectedDialog === "new-workflow";

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <NewFileButton setSelectedDialog={setSelectedDialog} />
      <DialogContent
        className={twMergeClsx(
          wideOverlay && "sm:max-w-xl md:max-w-2xl lg:max-w-3xl"
        )}
      >
        {selectedDialog === "new-dir" && (
          <NewDirectory
            path={data?.path}
            unallowedNames={
              data?.children?.map((file) => getFilenameFromPath(file.path)) ||
              []
            }
            close={() => setDialogOpen(false)}
          />
        )}
        {selectedDialog === "new-workflow" && (
          <NewWorkflow
            path={data?.path}
            unallowedNames={
              data?.children?.map((file) => getFilenameFromPath(file.path)) ||
              []
            }
            close={() => setDialogOpen(false)}
          />
        )}
        {selectedDialog === "new-service" && (
          <NewService
            path={data?.path}
            unallowedNames={
              data?.children?.map((file) => getFilenameFromPath(file.path)) ||
              []
            }
            close={() => setDialogOpen(false)}
          />
        )}
        {selectedDialog === "new-route" && (
          <NewRoute
            path={data?.path}
            unallowedNames={
              data?.children?.map((file) => getFilenameFromPath(file.path)) ||
              []
            }
            close={() => setDialogOpen(false)}
          />
        )}
        {selectedDialog === "new-consumer" && (
          <NewConsumer
            path={data?.path}
            unallowedNames={
              data?.children?.map((file) => getFilenameFromPath(file.path)) ||
              []
            }
            close={() => setDialogOpen(false)}
          />
        )}
      </DialogContent>
    </Dialog>
  );
};
