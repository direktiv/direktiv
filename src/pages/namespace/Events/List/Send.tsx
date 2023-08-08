import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { Calendar } from "lucide-react";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { exampleEvent } from "./exampleEvent";
import { useTranslation } from "react-i18next";

const Send = () => {
  const { t } = useTranslation();
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="primary" data-testid="event-create">
          {t("pages.events.list.send.dialogTrigger")}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            <Calendar />
            {t("pages.events.list.send.dialogHeader")}
          </DialogTitle>
        </DialogHeader>
        <Card
          className="grow p-4 pl-0"
          background="weight-1"
          data-testid="variable-create-card"
        >
          <div className="h-[500px]">
            <Editor value={exampleEvent} />
          </div>
        </Card>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button
            data-testid="variable-create-submit"
            type="submit"
            variant="primary"
          >
            {t("pages.events.list.send.submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default Send;
