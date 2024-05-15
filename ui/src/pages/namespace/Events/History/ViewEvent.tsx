import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { EventSchemaType, NewEventSchemaType } from "~/api/events/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Radio } from "lucide-react";
import { prettifyJsonString } from "~/util/helpers";
import { useReplayEvent } from "~/api/events/mutate/replayEvent";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const ViewEvent = ({
  event,
  handleOpenChange,
}: {
  event: EventSchemaType;
  handleOpenChange: (value: boolean) => void;
}) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const { mutate: replayEvent } = useReplayEvent({
    onSuccess: () => handleOpenChange(false),
  });

  const onSubmit: SubmitHandler<NewEventSchemaType> = () => {
    replayEvent(event.id);
  };

  const { handleSubmit } = useForm<NewEventSchemaType>({});

  const eventDataPretty = prettifyJsonString(JSON.stringify(event));

  return (
    <form
      id="send-event"
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col space-y-5"
    >
      <DialogHeader>
        <DialogTitle>
          <Radio />
          {t("pages.events.history.view.dialogHeader")}
        </DialogTitle>
      </DialogHeader>
      <Alert variant="info" className="text-sm">
        {t("pages.events.history.view.info")}
      </Alert>
      <Card
        className="grow p-4 pl-0"
        background="weight-1"
        data-testid="event-view-card"
      >
        <div className="h-[500px]">
          <Editor
            value={eventDataPretty}
            language="json"
            theme={theme ?? undefined}
            options={{ readOnly: true }}
          />
        </div>
      </Card>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button data-testid="event-view-submit" type="submit" variant="primary">
          {t("pages.events.history.view.submit")}
        </Button>
      </DialogFooter>
    </form>
  );
};

export default ViewEvent;
